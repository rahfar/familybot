package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Weather struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		Temp      float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		}
		string `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Day struct {
				Maxtemp_c float64 `json:"maxtemp_c"`
				Mintemp_c float64 `json:"mintemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func GetWeather(apikey string, cities []string) []Weather {
	weather := make([]Weather, 0)
	for _, city := range cities {
		w, err := callCurrentApi(apikey, city)
		if err != nil {
			log.Printf("[WARN] Could not get weather for %s: %v\n", city, err)
		} else {
			weather = append(weather, *w)
		}
	}
	return weather
}

func callCurrentApi(apikey string, city string) (*Weather, error) {
	base_url := "https://api.weatherapi.com/v1/forecast.json"
	query_str := fmt.Sprintf("?key=%s&q=%s&lang=ru&days=1&aqi=no&alerts=no", apikey, url.QueryEscape(city))
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
	var w Weather
	err = json.Unmarshal(body, &w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
