package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

type WeatherAPI struct {
	ApiKey     string
	Config     WeatherAPIConfig
	HttpClient *http.Client
}

type WeatherAPIConfig struct {
	Cities map[string]CityPosition `json:"cities"`
}

type CityPosition struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
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

func readConfigFile(configFilePath string) (WeatherAPIConfig, error) {
	var config WeatherAPIConfig

	configFile, err := os.Open(configFilePath)

	if err != nil {
		slog.Warn("Error opening config file", "err", err)
		return config, err
	}

	defer configFile.Close()

	configFileBytes, err := io.ReadAll(configFile)

	if err != nil {
		slog.Warn("Error reading config file", "err", err)
		return config, err
	}

	err = json.Unmarshal(configFileBytes, &config)

	if err != nil {
		slog.Warn("Error parsing config file", "err", err)
		return config, err
	}

	return config, nil
}

func NewWeatherAPI(apiKey string, configFile string, httpClient *http.Client) *WeatherAPI {
	cfg, err := readConfigFile(configFile)

	if err != nil {
		slog.Warn("Error reading config file", "err", err)
	}

	return &WeatherAPI{
		ApiKey:     apiKey,
		Config:     cfg,
		HttpClient: httpClient,
	}
}

func (w *WeatherAPI) GetWeather() []Weather {
	weather := make([]Weather, 0)

	for c, cp := range w.Config.Cities {
		w, err := w.callCurrentAPI(fmt.Sprintf("%f,%f", cp.Lat, cp.Lon))
		if err != nil {
			slog.Warn("could not get weather", "city", c, "err", err)
		} else {
			w.Location.Name = c
			weather = append(weather, *w)
		}
	}
	return weather
}

func (w *WeatherAPI) callCurrentAPI(latlon string) (*Weather, error) {
	const maxRetry = 3
	baseURL := "https://api.weatherapi.com/v1/forecast.json"
	queryStr := fmt.Sprintf("?key=%s&q=%s&lang=ru&days=1&aqi=no&alerts=no", w.ApiKey, url.QueryEscape(latlon))

	for i := 1; i <= maxRetry; i++ {
		resp, err := w.HttpClient.Get(baseURL + queryStr)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode/100 == 2 {
			var weather Weather
			if err := json.Unmarshal(body, &weather); err != nil {
				return nil, err
			}
			return &weather, nil
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
