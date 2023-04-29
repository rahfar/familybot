package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

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

func GetExchangeRates(apikey string) (*ExchangeRates, error) {
	base_url := "https://api.currencyapi.com/v3/latest"
	query_str := fmt.Sprintf("?apikey=%s", apikey)
	resp, err := http.Get(base_url + query_str)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non 2** HTTP status code: %d - %s - %s", resp.StatusCode, resp.Status, string(body))
	}
	var xr ExchangeRates
	err = json.Unmarshal(body, &xr)
	if err != nil {
		log.Printf("[ERROR] Could not unmarshal json body: %v", err)
		return nil, err
	}
	return &xr, nil
}

func GetHistoryExchangeRates(apikey string, datetime time.Time) (*ExchangeRates, error) {
	base_url := "https://api.currencyapi.com/v3/historical"
	query_str := fmt.Sprintf("?apikey=%s&date=%s", apikey, datetime.Format("2006-01-02"))
	resp, err := http.Get(base_url + query_str)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non 2** HTTP status code: %d - %s - %s", resp.StatusCode, resp.Status, string(body))
	}
	var xr ExchangeRates
	err = json.Unmarshal(body, &xr)
	if err != nil {
		log.Printf("[ERROR] Could not unmarshal json body: %v", err)
		return nil, err
	}
	return &xr, nil
}
