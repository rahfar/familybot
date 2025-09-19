package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	Name        string
	Description string
	Handler     func(*Bot, *tgbotapi.Message)
	Hidden      bool
}

var Commands = map[string]Command{
	"/ping": {
		Name:        "/ping",
		Description: "Провекра связи.",
		Handler:     ping,
		Hidden:      true,
	},
	"/whoami": {
		Name:        "/whoami",
		Description: "Возвращает chat_id и user_id",
		Handler:     whoAmI,
		Hidden:      true,
	},
	"/mourning": {
		Name:        "/mourning",
		Description: "Debug mourning job",
		Handler:     sendMourningDigest,
		Hidden:      true,
	},
	"/revision": {
		Name:        "/revision",
		Description: "Версия бота.",
		Handler:     getRevision,
		Hidden:      true,
	},
	"/weather": {
		Name:        "/weather",
		Description: "Прогноз погоды в заданных городах.",
		Handler:     getCurrentWeather,
		Hidden:      true,
	},
	"/restart": {
		Name:        "/restart",
		Description: "Сбросить контекст в работе с ChatGPT.",
		Handler:     restartChatGPT,
	},
	"/gpt": {
		Name:        "/gpt",
		Description: "Спросить ChatGPT.",
		Handler:     askChatGPT,
		Hidden:      true,
	},
	"/fix": {
		Name:        "/fix",
		Description: "Проверить и поправить грамматику в английском тексте.",
		Handler:     correctEnglish,
		Hidden:      true,
	},
	"/en2ru": {
		Name:        "/en2ru",
		Description: "Перевод с английского на русский.",
		Handler:     translateEng2Ru,
		Hidden:      true,
	},
	"/ru2en": {
		Name:        "/ru2en",
		Description: "Перевод с русского на английский.",
		Handler:     translateRu2Eng,
		Hidden:      true,
	},
	"/add": {
		Name:        "/add",
		Description: "Добавить чат (только для админов).",
		Handler:     addUser,
		Hidden:      true,
	},
	"/remove": {
		Name:        "/remove",
		Description: "Удалить чат (только для админов).",
		Handler:     removeUser,
		Hidden:      true,
	},
	"/users": {
		Name:        "/users",
		Description: "Список авторизованных чатов (только для админов).",
		Handler:     listUsers,
		Hidden:      true,
	},
	"/invite": {
		Name:        "/invite",
		Description: "Сгенерировать ссылку приглашения (только для админов).",
		Handler:     generateInvite,
		Hidden:      true,
	},
	"/start": {
		Name:        "/start",
		Description: "Начать работу с ботом.",
		Handler:     handleStartCommand,
		Hidden:      true,
	},
}

// Register commands that reference Commands itself in init to avoid initialization cycle.
func AddListCommand() {
	Commands["/list"] = Command{
		Name:        "/list",
		Description: "Список доступных команд.",
		Handler:     listCommands,
		Hidden:      true,
	}
}
