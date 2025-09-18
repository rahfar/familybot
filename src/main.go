package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jessevdk/go-flags"

	"github.com/rahfar/familybot/src/apiclient"
	"github.com/rahfar/familybot/src/bot"
	"github.com/rahfar/familybot/src/db"
)

var opts struct {
	Telegram struct {
		Token        string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID      int64  `long:"group" env:"GROUPID" description:"group id"`
		AdminUserIDs string `long:"adminuserids" env:"ADMINUSERIDS" description:"comma-separated list of admin user IDs" default:""`
	} `group:"telegram" namespace:"telegram" env-namespace:"TG"`
	WeatherAPI struct {
		Key        string `long:"key" env:"KEY"`
		ConfigFile string `long:"configfile" env:"configfile" default:"weatherapi_config.json" description:"config file for weather api"`
	} `group:"weatherapi" namespace:"weatherapi" env-namespace:"WEATHERAPI"`
	CurrencyAPI struct {
		Key string `long:"key" env:"KEY"`
	} `group:"currencyapi" namespace:"currencyapi" env-namespace:"CURRENCYAPI"`
	OpenaiAPI struct {
		Key string `long:"key" env:"KEY"`
	} `group:"openaiapi" namespace:"openaiapi" env-namespace:"OPENAIAPI"`
	MinifluxAPI struct {
		Key     string `long:"key" env:"KEY"`
		BaseURL string `long:"baseurl" env:"BASEURL"`
	} `group:"minifluxapi" namespace:"minifluxapi" env-namespace:"MINIFLUXAPI"`
	DeeplAPI struct {
		Key     string `long:"key" env:"KEY"`
		BaseURL string `long:"baseurl" env:"BASEURL" default:"https://api-free.deepl.com"`
	} `group:"deeplapi" namespace:"deeplapi" env-namespace:"DEEPLAPI"`
	RedisAddr string `long:"redisaddr" env:"REDIS_ADDR" default:"localhost:6379"`
	Host      string `long:"host" env:"HOST" default:"0.0.0.0"`
	Port      string `long:"port" env:"PORT" default:"8080"`
	Dbg       bool   `long:"debug" env:"DEBUG" description:"debug mode"`
}

func ConvertCommaSeparatedStringToInt64Slice(input string) ([]int64, error) {
	// Split the input string by commas
	parts := strings.Split(input, ",")
	intSlice := make([]int64, 0, len(parts))

	for _, part := range parts {
		// Trim whitespace from each part
		part = strings.TrimSpace(part)

		// Convert the string to int64
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting %s to int64: %v", part, err)
		}
		intSlice = append(intSlice, num)
	}

	return intSlice, nil
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		slog.Error("Error parsing options")
		panic(err)
	}
	logLevel := slog.LevelInfo
	if opts.Dbg {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: logLevel}))
	slog.SetDefault(logger)
	tgbotapi.SetLogger(log.Default())

	bot_api, err := tgbotapi.NewBotAPI(opts.Telegram.Token)
	if err != nil {
		slog.Error("could not initialize bot", "err", err)
		panic(err)
	}

	bot_api.Debug = opts.Dbg

	slog.Info("bot is authorized", "bot-username", bot_api.Self.UserName)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	dbClient := db.NewClient(opts.RedisAddr)

	exchangeAPI := &apiclient.ExchangeAPI{
		ApiKey:     opts.CurrencyAPI.Key,
		DBClient:   dbClient,
		HttpClient: httpClient,
	}
	openaiAPI := &apiclient.OpenaiAPI{
		ApiKey:     opts.OpenaiAPI.Key,
		HttpClient: httpClient,
	}
	deeplAPI := &apiclient.DeeplAPI{
		HttpClient: httpClient,
		DBClient:   dbClient,
		ApiKey:     opts.DeeplAPI.Key,
		BaseURL:    opts.DeeplAPI.BaseURL,
	}
	minifluxAPI := &apiclient.MinifluxAPI{
		ApiKey:  opts.MinifluxAPI.Key,
		BaseURL: opts.MinifluxAPI.BaseURL,
	}
	weatherAPI := apiclient.NewWeatherAPI(opts.WeatherAPI.Key, opts.WeatherAPI.ConfigFile, httpClient, dbClient)

	adminUserIDs, err := ConvertCommaSeparatedStringToInt64Slice(opts.Telegram.AdminUserIDs)
	if err != nil {
		slog.Error("Error parsing AdminUserIDs")
		panic(err)
	}

	b := bot.Bot{
		Token:        opts.Telegram.Token,
		Dbg:          opts.Dbg,
		Host:         opts.Host,
		Port:         opts.Port,
		AdminUserIDs: adminUserIDs,
		GroupID:      opts.Telegram.GroupID,
		ExchangeAPI:  exchangeAPI,
		OpenaiAPI:    openaiAPI,
		WeatherAPI:   weatherAPI,
		TGBotAPI:     bot_api,
		MinifluxAPI:  minifluxAPI,
		DeeplAPI:     deeplAPI,
		DBClient:     dbClient,
	}

	b.Run()
}
