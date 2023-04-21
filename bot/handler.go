package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func ping(message tgbotapi.Message) string {
	return "понг"
}
