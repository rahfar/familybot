package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	Names       []string
	Description string
	Handler     func(*Bot, *tgbotapi.Message) tgbotapi.MessageConfig
}

func findCommand(commands []Command, name string) *Command {
	for i, cmd := range commands {
		for _, n := range cmd.Names {
			if strings.EqualFold(n, name) {
				return &commands[i]
			}
		}
	}
	return nil
}

var Commands = []Command{
	{
		Names:       []string{"!пинг", "!ping"},
		Description: "Провекра связи.",
		Handler:     ping,
	},
	{
		Names:       []string{"!погода", "!weather"},
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
	},
	{
		Names:       []string{"!чат", "!gpt"},
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
	},
	{
		Names:       []string{"!продажи", "!sales"},
		Description: "Текущие продажи.",
		Handler:     getYesterdaySales,
	},
	{
		Names:       []string{"!анекдот", "!joke"},
		Description: "Свежий анекдот (может быть даже смешной).",
		Handler:     getAnecdote,
	},
	{
		Names:       []string{"!новости", "!news"},
		Description: "Последние новости с сайта Коммерсант.",
		Handler:     getLatestNews,
	},
}
