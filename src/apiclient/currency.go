package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rahfar/familybot/src/db"
)

type ExchangeAPI struct {
	ApiKey      string
	HttpClient  *http.Client
	DBClient *db.Client
}

type ExchangeRates struct {
	Meta struct {
		Update_time time.Time `json:"last_updated_at"`
	} `json:"meta"`
	Data struct {
		BTC struct {
			Value float64 `json:"value"`
		} `json:"BTC"`
		EUR struct {
			Value float64 `json:"value"`
		} `json:"EUR"`
		RUB struct {
			Value float64 `json:"value"`
		} `json:"RUB"`
	} `json:"data"`
}

func (e *ExchangeAPI) GetExchangeRates(datetime time.Time) (*ExchangeRates, error) {
	const maxRetry = 3
	var xr ExchangeRates
	var baseURL, queryStr string
	ctx := context.Background()

	if datetime.Before(time.Now().Add(-24 * time.Hour)) {
		baseURL = "https://api.currencyapi.com/v3/historical"
		queryStr = fmt.Sprintf("?apikey=%s&date=%s", e.ApiKey, datetime.Format("2006-01-02"))
	} else {
		baseURL = "https://api.currencyapi.com/v3/latest"
		queryStr = fmt.Sprintf("?apikey=%s", e.ApiKey)
	}

	v, err := e.DBClient.GetCurrencyRates(ctx, datetime)
	if err == nil {
		slog.Info("hit currencyapi cache", "key", e.DBClient.CurrencyKey(datetime))
		err := json.Unmarshal([]byte(v), &xr)
		if err == nil {
			return &xr, nil
		}
	}

	for i := 1; i <= maxRetry; i++ {
		resp, err := e.HttpClient.Get(baseURL + queryStr)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode/100 == 2 {
			if err := e.DBClient.SetCurrencyRates(ctx, datetime, body); err != nil {
				slog.Info("could not write cache", "err", err)
			}
			err := json.Unmarshal(body, &xr)
			if err != nil {
				slog.Error("could not unmarshal json body", "err", err)
				return nil, err
			}
			return &xr, nil
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(5 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return nil, fmt.Errorf("max retries reached")
}
