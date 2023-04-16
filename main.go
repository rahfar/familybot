package main

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/rahfar/familybot/bot"
)

var opts struct {
	Telegram struct {
		Token   string        `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		Group   string        `long:"group" env:"GROUP" description:"group name/id" default:"test"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`

	Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal("Error parsing options")
	}
	bot.RunBot(opts.Telegram.Token, opts.Dbg)
}
