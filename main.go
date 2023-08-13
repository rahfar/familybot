package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jessevdk/go-flags"

	"github.com/rahfar/familybot/bot"
	"github.com/rahfar/familybot/bot/apiclient"
)

var opts struct {
	Telegram struct {
		Token            string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID          int64  `long:"group" env:"GROUP" description:"group id"`
		AllowedUsernames string `long:"allowedusernames" env:"ALLOWEDUSERNAMES" description:"list of usernames that will have access to the bot" default:""`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	WeatherAPI struct {
		Key    string `long:"key" env:"KEY"`
		Cities string `long:"cities" env:"CITIES"`
	} `group:"weatherapi" namespace:"weatherapi" env-namespace:"WEATHERAPI"`
	CurrencyAPI struct {
		Key string `long:"key" env:"KEY"`
	} `group:"currencyapi" namespace:"currencyapi" env-namespace:"CURRENCYAPI"`
	OpenaiAPI struct {
		Key string `long:"key" env:"KEY"`
	} `group:"openaiapi" namespace:"openaiapi" env-namespace:"OPENAIAPI"`
	GoogleSheetsAPI struct {
		Key           string `long:"key" env:"KEY"`
		SpreadSheetID string `long:"spreadsheetid" env:"SPREADSHEETID"`
	} `group:"googlesheetsapi" namespace:"googlesheetsapi" env-namespace:"GOOGLESHEETSAPI"`
	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

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

	httpClient := &http.Client{Timeout: 15 * time.Second}
	openaiHttpClient := &http.Client{Timeout: 60 * time.Second}

	anekdotAPI := &apiclient.AnecdoteAPI{HttpClient: httpClient}
	exchangeAPI := &apiclient.ExchangeAPI{ApiKey: opts.CurrencyAPI.Key, HttpClient: httpClient}
	sheetsAPI := &apiclient.SheetsAPI{ApiKey: opts.GoogleSheetsAPI.Key, SpreadsheetId: opts.GoogleSheetsAPI.SpreadSheetID}
	kommerstantAPI := &apiclient.KommersantAPI{HttpClient: httpClient}
	openaiAPI := &apiclient.OpenaiAPI{ApiKey: opts.OpenaiAPI.Key, HttpClient: openaiHttpClient}
	weatherAPI := &apiclient.WeatherAPI{ApiKey: opts.WeatherAPI.Key, Cities: opts.WeatherAPI.Cities, HttpClient: httpClient}

	b := bot.Bot{
		Token:            opts.Telegram.Token,
		Dbg:              opts.Dbg,
		AllowedUsernames: strings.Split(opts.Telegram.AllowedUsernames, ","),
		GroupID:          opts.Telegram.GroupID,
		Commands:         bot.Commands,
		AnekdotAPI:       anekdotAPI,
		ExchangeAPI:      exchangeAPI,
		SheetsAPI:        sheetsAPI,
		KommersantAPI:    kommerstantAPI,
		OpenaiAPI:        openaiAPI,
		WeatherAPI:       weatherAPI,
		TGBotAPI:         bot_api,
	}
	b.Run()
}
