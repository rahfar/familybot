package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jessevdk/go-flags"

	"github.com/rahfar/familybot/bot"
	"github.com/rahfar/familybot/bot/apiclient"
)

var opts struct {
	Telegram struct {
		Token   string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID int64  `long:"group" env:"GROUP" description:"group id"`
		Chats   string `long:"chats" env:"CHATS" description:"acceptable usernames" default:""`
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
	DataDir string `long:"data_dir" env:"DATA_DIR" description:"path to data directory" default:"./data"`
	Dbg     bool   `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal("[ERROR] Error parsing options")
	}

	bot_api, err := tgbotapi.NewBotAPI(opts.Telegram.Token)
	if err != nil {
		log.Panic(err)
	}

	bot_api.Debug = opts.Dbg

	log.Printf("[INFO] Authorized on account %s", bot_api.Self.UserName)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	openaiHttpClient := &http.Client{Timeout: 60 * time.Second}

	anekdotAPI := &apiclient.AnecdoteAPI{HttpClient: httpClient}
	exchangeAPI := &apiclient.ExchangeAPI{ApiKey: opts.CurrencyAPI.Key, HttpClient: httpClient}
	sheetsAPI := &apiclient.SheetsAPI{ApiKey: opts.GoogleSheetsAPI.Key, SpreadsheetId: opts.GoogleSheetsAPI.SpreadSheetID}
	kommerstantAPI := &apiclient.KommersantAPI{HttpClient: httpClient}
	openaiAPI := &apiclient.OpenaiAPI{ApiKey: opts.OpenaiAPI.Key, HttpClient: openaiHttpClient}
	weatherAPI := &apiclient.WeatherAPI{ApiKey: opts.WeatherAPI.Key, Cities: opts.WeatherAPI.Cities, HttpClient: httpClient}

	bot := bot.Bot{
		Token:         opts.Telegram.Token,
		Dbg:           opts.Dbg,
		GroupID:       opts.Telegram.GroupID,
		Chats:         strings.Split(opts.Telegram.Chats, ","),
		DataDir:       opts.DataDir,
		AnekdotAPI:    anekdotAPI,
		ExchangeAPI:   exchangeAPI,
		SheetsAPI:     sheetsAPI,
		KommersantAPI: kommerstantAPI,
		OpenaiAPI:     openaiAPI,
		WeatherAPI:    weatherAPI,
		TGBotAPI:      bot_api,
	}
	bot.Run()
}
