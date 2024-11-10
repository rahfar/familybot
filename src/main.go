package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jessevdk/go-flags"
	"github.com/redis/go-redis/v9"

	"github.com/rahfar/familybot/src/apiclient"
	"github.com/rahfar/familybot/src/bot"
)

var opts struct {
	Telegram struct {
		Token            string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID          int64  `long:"group" env:"GROUP" description:"group id"`
		AllowedUsernames string `long:"allowedusernames" env:"ALLOWEDUSERNAMES" description:"list of usernames that will have access to the bot" default:""`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	WeatherAPI struct {
		Key        string `long:"key" env:"KEY"`
		ConfigFile string `long:"configfile" env:"configfile" default:"weatherapi_config.json" description:"config file for weather api"`
	} `group:"weatherapi" namespace:"weatherapi" env-namespace:"WEATHERAPI"`
	CurrencyAPI struct {
		Key string `long:"key" env:"KEY"`
	} `group:"currencyapi" namespace:"currencyapi" env-namespace:"CURRENCYAPI"`
	OpenaiAPI struct {
		Key      string `long:"key" env:"KEY"`
		GPTModel string `long:"gptmodel" env:"GPTMODEL" default:"gpt-3.5-turbo"`
	} `group:"openaiapi" namespace:"openaiapi" env-namespace:"OPENAIAPI"`
	AnthropicAPI struct {
		Key        string `long:"key" env:"KEY"`
		Model      string `long:"model" env:"MODEL" default:"claude-3-5-sonnet-latest"`
		ApiVersion string `long:"apiversion" env:"APIVERSION" default:"2023-06-01"`
		MaxTokens  int    `long:"maxtokens" env:"MAXTOKENS" default:"1024"`
	} `group:"anthropicapi" namespace:"anthropicapi" env-namespace:"ANTHROPICAPI"`
	MinifluxAPI struct {
		Key     string `long:"key" env:"KEY"`
		BaseURL string `long:"baseurl" env:"BASEURL"`
		SiteURL string `long:"siteurl" env:"SITEURL"`
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)
	tgbotapi.SetLogger(log.Default())

	if _, err := flags.Parse(&opts); err != nil {
		slog.Error("Error parsing options")
		panic(err)
	}

	bot_api, err := tgbotapi.NewBotAPI(opts.Telegram.Token)
	if err != nil {
		slog.Error("could not initialize bot", "err", err)
		panic(err)
	}

	bot_api.Debug = opts.Dbg

	slog.Info("bot is authorized", "bot-username", bot_api.Self.UserName)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	redisClient := redis.NewClient(&redis.Options{Addr: opts.RedisAddr})

	exchangeAPI := &apiclient.ExchangeAPI{ApiKey: opts.CurrencyAPI.Key, RedisClient: redisClient, HttpClient: httpClient}
	openaiAPI := &apiclient.OpenaiAPI{ApiKey: opts.OpenaiAPI.Key, HttpClient: httpClient, GPTModel: opts.OpenaiAPI.GPTModel}
	deeplAPI := &apiclient.DeeplAPI{HttpClient: httpClient, RedisClient: redisClient, ApiKey: opts.DeeplAPI.Key, BaseURL: opts.DeeplAPI.BaseURL}
	minifluxAPI := &apiclient.MinifluxAPI{ApiKey: opts.MinifluxAPI.Key, BaseURL: opts.MinifluxAPI.BaseURL, SiteURL: opts.MinifluxAPI.SiteURL}
	weatherAPI := apiclient.NewWeatherAPI(opts.WeatherAPI.Key, opts.WeatherAPI.ConfigFile, httpClient, redisClient)
	anthropicAPI := &apiclient.AnthropicAPI{
		ApiKey:      opts.AnthropicAPI.Key,
		HttpClient:  httpClient,
		RedisClient: redisClient,
		Model:       opts.AnthropicAPI.Model,
		ApiVersion:  opts.AnthropicAPI.ApiVersion,
		MaxTokens:   opts.AnthropicAPI.MaxTokens,
	}

	b := bot.Bot{
		Token:            opts.Telegram.Token,
		Dbg:              opts.Dbg,
		Host:             opts.Host,
		Port:             opts.Port,
		AllowedUsernames: strings.Split(opts.Telegram.AllowedUsernames, ","),
		GroupID:          opts.Telegram.GroupID,
		Commands:         bot.Commands,
		AskGPTCache:      expirable.NewLRU[string, []apiclient.GPTResponse](1000, nil, time.Minute*30),
		ExchangeAPI:      exchangeAPI,
		OpenaiAPI:        openaiAPI,
		WeatherAPI:       weatherAPI,
		TGBotAPI:         bot_api,
		MinifluxAPI:      minifluxAPI,
		DeeplAPI:         deeplAPI,
		AnthropicAPI:     anthropicAPI,
	}
	b.Run()
}
