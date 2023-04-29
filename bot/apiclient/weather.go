package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type WeatherData struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	CurrentWeather struct {
		Temp      float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		}
		string `json:"condition"`
	} `json:"current"`
}

func Get_weather(apikey string, cities []string) []WeatherData {
	weather := make([]WeatherData, 0)
	for _, city := range cities {
		w, err := api_call(apikey, city)
		if err != nil {
			log.Printf("[ERROR] Could not get weather for %s: %v\n", city, err)
		} else {
			weather = append(weather, *w)
		}
	}
	return weather
}

func api_call(apikey string, city string) (*WeatherData, error) {
	base_url := "https://api.weatherapi.com/v1/current.json"
	query_str := fmt.Sprintf("?key=%s&q=%s&lang=ru", apikey, url.QueryEscape(city))
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
	var w WeatherData
	err = json.Unmarshal(body, &w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

