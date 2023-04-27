package api_clients

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Currency struct {
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

func Get_currency(apikey string) (*Currency, error) {
	base_url := "https://api.currencyapi.com/v3/latest"
	query_str := fmt.Sprintf("?apikey=%s", apikey)
	resp, err := http.Get(base_url + query_str)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non 2** HTTP status code: %d - %s", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var c Currency
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Fatalln(err)
	}
	return &c, nil
}
