package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	Name        string
	Description string
	Handler     func(*Bot, *tgbotapi.Message) tgbotapi.MessageConfig
}

func findCommand(commands []Command, name string) *Command {
	for i := range commands {
		if strings.EqualFold(commands[i].Name, name) {
			return &commands[i]
		}
	}
	return nil
}

var Commands = []Command{
	{
		Name:        "!пинг",
		Description: "Провекра связи.",
		Handler:     ping,
	},
	{
		Name:        "!погода",
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
	},
	{
		Name:        "!чат",
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
	},
	{
		Name:        "!продажи",
		Description: "Текущие продажи.",
		Handler:     getYesterdaySales,
	},
	{
		Name:        "!анекдот",
		Description: "Свежий анекдот (может быть даже смешной).",
		Handler:     getAnecdote,
	},
	{
		Name:        "!новости",
		Description: "Последние новости с сайта Коммерсант.",
		Handler:     getLatestNews,
	},
}
