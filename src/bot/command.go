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
	"/whoami": {
		Name:        "/whoami",
		Description: "Возвращает chat_id и user_id",
		Handler:     whoAmI,
	},
	"/mourning": {
		Name:        "/mourning",
		Description: "Debug mourning job",
		Handler:     sendMourningDigest,
	},
	"/revision": {
		Name:        "/revision",
		Description: "Версия бота.",
		Handler:     getRevision,
	},
	"/weather": {
		Name:        "/weather",
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
	},
	"/new": {
		Name:        "/new",
		Description: "Сбросить контекст в работе с ChatGPT.",
		Handler:     newChatGPT,
	},
	"/gpt": {
		Name:        "/gpt",
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
	},
	"/check_eng": {
		Name:        "/check_eng",
		Description: "Проверить и поправить грамматику в английском тексте.",
		Handler:     correctEnglish,
	},
	"/en2ru": {
		Name:        "/en2ru",
		Description: "Перевод с английского на русский.",
		Handler:     translateEng2Ru,
	},
	"/ru2en": {
		Name:        "/ru2en",
		Description: "Перевод с русского на английский.",
		Handler:     translateRu2Eng,
	},
}
