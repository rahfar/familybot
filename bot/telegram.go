package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func RunBot(token string, dbg bool) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = dbg

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		resp := ping(update.Message.Text)
		if resp == "" {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp)

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
