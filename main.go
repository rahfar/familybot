package main

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/rahfar/familybot/bot"
)

var opts struct {
	Telegram struct {
		Token   string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		GroupID int64  `long:"group" env:"GROUP" description:"group id" default:"0"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	DataDir string `long:"data_dir" env:"DATA_DIR" description:"path to data directory" default:"./data"`
	Dbg     bool   `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal("Error parsing options")
	}

	bot := bot.Bot{
		Token:   opts.Telegram.Token,
		Dbg:     opts.Dbg,
		GroupID: opts.Telegram.GroupID,
		DataDir: opts.DataDir,
	}
	bot.Run()
}
