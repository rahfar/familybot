package bot

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Token   string
	Dbg     bool
	GroupID int64
	DataDir string
}

func (b *Bot) Run() {
	bot_api, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		log.Panic(err)
	}

	bot_api.Debug = b.Dbg

	log.Printf("Authorized on account %s", bot_api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot_api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		if update.Message.Chat == nil {
			continue
		}
		if b.GroupID != update.Message.Chat.ID {
			log.Printf("Skip message from unsupported chat. Chat: %+v\n", *update.Message.Chat)
			continue
		}
		b.on_message(*update.Message, bot_api)
	}
}

func (b *Bot) on_message(message tgbotapi.Message, bot_api *tgbotapi.BotAPI) {
	var resp string
	switch {
	case strings.HasPrefix(message.Text, "пинг"):
		resp = ping(message)
	case strings.HasPrefix(message.Text, "!время"):
		resp = get_users_current_time(b.DataDir)
	case message.Location != nil && message.From != nil:
		remember_tz(message, b.DataDir)
		return
	default:
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, resp)

	if _, err := bot_api.Send(msg); err != nil {
		log.Panic(err)
	}
}
