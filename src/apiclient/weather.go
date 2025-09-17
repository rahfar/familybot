package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rahfar/familybot/src/db"
)

type WeatherAPI struct {
	ApiKey      string
	Config      WeatherAPIConfig
	HttpClient  *http.Client
	DBClient *db.Client
}

type WeatherAPIConfig struct {
	Cities map[string]CityPosition `json:"cities"`
}

type CityPosition struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Root struct represents the entire JSON response
type WeatherResponse struct {
	Cod     string        `json:"cod"`
	Message int           `json:"message"`
	Cnt     int           `json:"cnt"`
	List    []WeatherItem `json:"list"`
	City    City          `json:"city"`
}

// WeatherItem struct represents each item in the 'list' array
type WeatherItem struct {
	Dt         int64     `json:"dt"`
	Main       Main      `json:"main"`
	Weather    []Weather `json:"weather"`
	Clouds     Clouds    `json:"clouds"`
	Wind       Wind      `json:"wind"`
	Visibility int       `json:"visibility"`
	Pop        float64   `json:"pop"`
	Sys        Sys       `json:"sys"`
	DtTxt      string    `json:"dt_txt"`
}

// Main struct contains details on temperature and pressure
type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
	TempKf    float64 `json:"temp_kf"`
}

// Weather struct provides weather summary and icon
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Clouds struct provides cloudiness percentage
type Clouds struct {
	All int `json:"all"`
}

// Wind struct contains information about wind speed and direction
type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

// Sys struct provides part of day information
type Sys struct {
	Pod string `json:"pod"`
}

// City struct contains city information
type City struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Coord      Coord  `json:"coord"`
	Country    string `json:"country"`
	Population int    `json:"population"`
	Timezone   int    `json:"timezone"`
	Sunrise    int64  `json:"sunrise"`
	Sunset     int64  `json:"sunset"`
}

// Coord struct provides geographical coordinates of the city
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
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

func NewWeatherAPI(apiKey string, configFile string, httpClient *http.Client, dbClient *db.Client) *WeatherAPI {
	cfg, err := readConfigFile(configFile)

	if err != nil {
		slog.Warn("Error reading config file", "err", err)
	}

	return &WeatherAPI{
		ApiKey:      apiKey,
		Config:      cfg,
		HttpClient:  httpClient,
		DBClient: dbClient,
	}
}

func (w *WeatherAPI) GetWeather() []WeatherResponse {
	weather := make([]WeatherResponse, 0)

	for c, cp := range w.Config.Cities {
		w, err := w.callCurrentAPI(cp.Lat, cp.Lon)
		if err != nil {
			slog.Warn("could not get weather", "city", c, "err", err)
		} else {
			w.City.Name = c
			weather = append(weather, *w)
		}
	}
	return weather
}

func (w *WeatherAPI) GetMinMaxTemp(weather WeatherResponse) (float64, float64) {
	var minTemp, maxTemp float64
	// Parse the date of the first entry
	firstDate, err := time.Parse("2006-01-02 15:04:05", weather.List[0].DtTxt)
	if err != nil {
		return minTemp, maxTemp
	}
	firstDay := firstDate.Day()

	maxTemp = weather.List[0].Main.Temp
	minTemp = weather.List[0].Main.Temp

	// Iterate over the weather list to find the maximum temperature for the first day
	for _, item := range weather.List {
		itemDate, err := time.Parse("2006-01-02 15:04:05", item.DtTxt)
		if err != nil {
			continue
		}
		if itemDate.Day() == firstDay {
			if item.Main.Temp > maxTemp {
				maxTemp = item.Main.Temp
			}
			if item.Main.Temp < minTemp {
				minTemp = item.Main.Temp
			}
		} else {
			break // Stop when leaving the first day
		}
	}

	return minTemp, maxTemp
}

func (w *WeatherAPI) callCurrentAPI(lat, lon float64) (*WeatherResponse, error) {
	const maxRetry = 3
	var weather WeatherResponse
	ctx := context.Background()
	baseURL := "https://api.openweathermap.org/data/2.5/forecast"
	queryStr := fmt.Sprintf("?lat=%f&lon=%f&appid=%s&lang=ru&units=metric", lat, lon, w.ApiKey)

	v, err := w.DBClient.GetWeatherData(ctx, lat, lon)
	if err == nil {
		slog.Info("hit weatherapi cache", "key", w.DBClient.WeatherKey(lat, lon))
		err := json.Unmarshal([]byte(v), &weather)
		if err == nil {
			return &weather, nil
		}
	}

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
			err := w.DBClient.SetWeatherData(ctx, lat, lon, body)
			if err != nil {
				slog.Info("could not write cache", "err", err)
			}
			err = json.Unmarshal(body, &weather)
			if err != nil {
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
