package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	Name        string
	Description string
	Handler     func(*Bot, *tgbotapi.Message)
}

var Commands = map[string]Command{
	"/ping": {
		Name:        "/ping",
		Description: "Провекра связи.",
		Handler:     ping,
	},
	"/weather": {
		Name:        "/weather",
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
	},
	"/gpt": {
		Name:        "/gpt",
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
	},
	"/revision": {
		Name:        "/revision",
		Description: "Версия бота.",
		Handler:     getRevision,
	},
	"/eng": {
		Name:        "/eng",
		Description: "Проверить и поправить грамматику в английском тексте.",
		Handler:     correctEnglish,
	},
	"/new": {
		Name:        "/new",
		Description: "Сбросить контекст в работе с ChatGPT.",
		Handler:     newChatGPT,
	},
	"/whoami": {
		Name:        "/whoami",
		Description: "Возвращает chat_id и user_id",
		Handler:     whoAmI,
	},
	"/mourning": {
		Name:        "/mourning",
		Description: "Debug mourning job",
		Handler:     mourningDebug,
	},
}
