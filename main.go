package main

import (
	"log"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/rahfar/familybot/bot"
)

var opts struct {
	Telegram struct {
		Token       string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID     int64  `long:"group" env:"GROUP" description:"group id" default:"0"`
		AdminChatID int64  `long:"admin_chat" env:"ADMIN_CHAT" description:"admin chat id" default:"0"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	WeatherAPI struct {
		Key    string   `long:"apikey" env:"KEY"`
		Cities string `long:"cities" env:"CITIES"`
	} `group:"weatherapi" namespace:"weatherapi" env-namespace:"WEATHERAPI"`
	CurrencyAPI struct {
		Key string `long:"apikey" env:"KEY"`
	} `group:"currencyapi" namespace:"currencyapi" env-namespace:"CURRENCYAPI"`
	DataDir string `long:"data_dir" env:"DATA_DIR" description:"path to data directory" default:"./data"`
	Dbg     bool   `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal("Error parsing options")
	}

	bot := bot.Bot{
		Token:            opts.Telegram.Token,
		Dbg:              opts.Dbg,
		GroupID:          opts.Telegram.GroupID,
		DataDir:          opts.DataDir,
		AdminChatID:      opts.Telegram.AdminChatID,
		WeatherAPIKey:    opts.WeatherAPI.Key,
		WeatherAPICities: strings.Split(opts.WeatherAPI.Cities, ","),
		CurrencyAPIKey:   opts.CurrencyAPI.Key,
	}
	bot.Run()
}
