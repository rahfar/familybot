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
	const max_retry int = 3
	base_url := "https://api.currencyapi.com/v3/latest"
	query_str := fmt.Sprintf("?apikey=%s", e.ApiKey)
	body := []byte{}

	for i := 1; i <= max_retry; i += 1 {
		resp, err := e.HttpClient.Get(base_url + query_str)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode/100 == 2 {
			break
		} else if i < max_retry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	var xr ExchangeRates
	err := json.Unmarshal(body, &xr)
	if err != nil {
		slog.Error("could not unmarshal json body", "err", err)
		return nil, err
	}
	return &xr, nil
}

func (e *ExchangeAPI) GetHistoryExchangeRates(datetime time.Time) (*ExchangeRates, error) {
	const max_retry int = 3
	base_url := "https://api.currencyapi.com/v3/historical"
	query_str := fmt.Sprintf("?apikey=%s&date=%s", e.ApiKey, datetime.Format("2006-01-02"))
	body := []byte{}

	for i := 1; i <= max_retry; i += 1 {
		resp, err := e.HttpClient.Get(base_url + query_str)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode/100 == 2 {
			break
		} else if i < max_retry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}
	var xr ExchangeRates
	err := json.Unmarshal(body, &xr)
	if err != nil {
		slog.Error("could not unmarshal json body", "err", err)
		return nil, err
	}
	return &xr, nil
}
