package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rahfar/familybot/bot/api_clients"
)

type Bot struct {
	Token            string
	Dbg              bool
	GroupID          int64
	AdminChatID      int64
	DataDir          string
	WeatherAPIKey    string
	WeatherAPICities []string
	CurrencyAPIKey   string
}

func (b *Bot) Run() {
	bot_api, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		log.Panic(err)
	}

	bot_api.Debug = b.Dbg

	log.Printf("Authorized on account %s", bot_api.Self.UserName)

	go b.mourning_job(bot_api)

	update_cfg := tgbotapi.NewUpdate(0)
	update_cfg.Timeout = 60

	updates := bot_api.GetUpdatesChan(update_cfg)

	for update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}
		if update.Message.Chat.ID != b.GroupID && update.Message.Chat.ID != b.AdminChatID {
			log.Printf("Skip message from unsupported chat. Chat: %+v\n", *update.Message.Chat)
			continue
		}
		b.on_message(*update.Message, bot_api)
	}
}

func (b *Bot) on_message(message tgbotapi.Message, bot_api *tgbotapi.BotAPI) {
	var resp string
	switch {
	case strings.HasPrefix(message.Text, "!пинг"):
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

func (b *Bot) mourning_job(bot_api *tgbotapi.BotAPI) {
	log.Println("Starting mourning job")
	for {
		text := "Доброе утро! 🌅\n"
		wait_untile_mourning()
		// call currency api
		c, err := api_clients.Get_currency(b.CurrencyAPIKey)
		if err != nil {
			log.Printf("[ERROR] Could not get currency exchange rates: %v", err)
		} else {
			text += fmt.Sprintf("\nКурсы валют:\n    USD %.2f₽\n    EUR %.2f₽\n    BTC %.2f$\n", c.Data.RUB.Value, c.Data.RUB.Value/c.Data.EUR.Value, 1.0/c.Data.BTC.Value)
		}
		// call weather api
		weather := api_clients.Get_weather(b.WeatherAPIKey, b.WeatherAPICities)
		if len(weather) > 0 {
			text += "\nПрогноз погоды:\n"
			for _, w := range weather {
				text += fmt.Sprintf("    %s: %.1f°C, %s\n", w.Location.Name, w.CurrentWeather.Temp, w.CurrentWeather.Condition.Text)
			}
		}
		// send message to group
		msg := tgbotapi.NewMessage(b.GroupID, text)

		if _, err := bot_api.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func wait_untile_mourning() {
	t := time.Now()
	var desiredTime time.Time
	if t.Hour() > 7 {
		desiredTime = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location()).Add(24 * time.Hour)
	} else {
		desiredTime = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())
	}
	time.Sleep(desiredTime.Sub(t))
}
