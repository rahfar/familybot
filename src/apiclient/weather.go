package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

type WeatherResponse struct {
	Coord      Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int       `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Clouds     Clouds    `json:"clouds"`
	Dt         int64     `json:"dt"`
	Sys        Sys       `json:"sys"`
	Timezone   int       `json:"timezone"`
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Cod        int       `json:"cod"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Type    int    `json:"type"`
	ID      int    `json:"id"`
	Country string `json:"country"`
	Sunrise int64  `json:"sunrise"`
	Sunset  int64  `json:"sunset"`
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

func (w *WeatherAPI) GetWeather() []WeatherResponse {
	weather := make([]WeatherResponse, 0)

	for c, cp := range w.Config.Cities {
		w, err := w.callCurrentAPI(cp.Lat, cp.Lon)
		if err != nil {
			slog.Warn("could not get weather", "city", c, "err", err)
		} else {
			w.Name = c
			weather = append(weather, *w)
		}
	}
	return weather
}

func (w *WeatherAPI) callCurrentAPI(lat, lon float64) (*WeatherResponse, error) {
	const maxRetry = 3
	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	queryStr := fmt.Sprintf("?lat=%f&lon=%f&appid=%s&lang=ru&units=metric", lat, lon, w.ApiKey)

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
			var weather WeatherResponse
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
