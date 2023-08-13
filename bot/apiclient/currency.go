package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type ExchangeAPI struct {
	ApiKey     string
	HttpClient *http.Client
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

func (e *ExchangeAPI) GetExchangeRates() (*ExchangeRates, error) {
	const maxRetry = 3
	baseURL := "https://api.currencyapi.com/v3/latest"
	queryStr := fmt.Sprintf("?apikey=%s", e.ApiKey)

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
			var xr ExchangeRates
			if err := json.Unmarshal(body, &xr); err != nil {
				slog.Error("could not unmarshal json body", "err", err)
				return nil, err
			}
			return &xr, nil
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return nil, fmt.Errorf("max retries reached")
}

func (e *ExchangeAPI) GetHistoryExchangeRates(datetime time.Time) (*ExchangeRates, error) {
	const maxRetry = 3
	baseURL := "https://api.currencyapi.com/v3/historical"
	queryStr := fmt.Sprintf("?apikey=%s&date=%s", e.ApiKey, datetime.Format("2006-01-02"))

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
			var xr ExchangeRates
			if err := json.Unmarshal(body, &xr); err != nil {
				slog.Error("could not unmarshal json body", "err", err)
				return nil, err
			}
			return &xr, nil
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return nil, fmt.Errorf("max retries reached")
}
