package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	Name        string
	Description string
	Handler     func(*Bot, *tgbotapi.Message)
}

func findCommand(commands []Command, name string) *Command {
	for i, cmd := range commands {
		if strings.EqualFold(cmd.Name, name) {
			return &commands[i]
		}
	}
	return nil
}

var Commands = []Command{
	{
		Name:        "/ping",
		Description: "Провекра связи.",
		Handler:     ping,
	},
	{
		Name:        "/weather",
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
	},
	{
		Name:        "/gpt",
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
	},
	{
		Name:        "/news",
		Description: "Последние новости с сайта Коммерсант.",
		Handler:     getLatestNews,
	},
	{
		Name:        "/revision",
		Description: "Версия бота.",
		Handler:     getRevision,
	},
	{
		Name:        "/eng",
		Description: "Проверить и поправить грамматику в английском тексте.",
		Handler:     correctEnglish,
	},
	{
		Name:        "/new",
		Description: "Сбросить контекст в работе с ChatGPT.",
		Handler:     newChatGPT,
	},
}
