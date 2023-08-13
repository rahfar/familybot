package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WeatherAPI struct {
	ApiKey     string
	Cities     string
	HttpClient *http.Client
}

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

func (w *WeatherAPI) GetWeather() []Weather {
	weather := make([]Weather, 0)
	cities := strings.Split(w.Cities, ",")
	for _, city := range cities {
		w, err := w.callCurrentApi(city)
		if err != nil {
			slog.Warn("could not get weather", "city", city, "err", err)
		} else {
			weather = append(weather, *w)
		}
	}
	return weather
}

func (w *WeatherAPI) callCurrentApi(city string) (*Weather, error) {
	const max_retry int = 3
	base_url := "https://api.weatherapi.com/v1/forecast.json"
	query_str := fmt.Sprintf("?key=%s&q=%s&lang=ru&days=1&aqi=no&alerts=no", w.ApiKey, url.QueryEscape(city))
	body := []byte{}

	for i := 1; i <= max_retry; i += 1 {
		resp, err := w.HttpClient.Get(base_url + query_str)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode/100 == 2 {
			break
		} else if i < max_retry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(5 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	var weather Weather
	err := json.Unmarshal(body, &weather)
	if err != nil {
		return nil, err
	}
	return &weather, nil
}
