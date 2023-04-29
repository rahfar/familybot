package bot

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rahfar/familybot/bot/apiclient"
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
	OpenaiAPIKey     string
}

func (b *Bot) Run() {
	bot_api, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		log.Panic(err)
	}

	bot_api.Debug = b.Dbg

	log.Printf("[INFO] Authorized on account %s", bot_api.Self.UserName)

	go b.mourningJob(bot_api)

	update_cfg := tgbotapi.NewUpdate(0)
	update_cfg.Timeout = 60

	updates := bot_api.GetUpdatesChan(update_cfg)

	for update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}
		if update.Message.Chat.ID != b.GroupID && update.Message.Chat.ID != b.AdminChatID {
			log.Printf("[INFO] Skip message from unsupported chat. Chat: %+v\n", *update.Message.Chat)
			continue
		}
		b.onMessage(*update.Message, bot_api)
	}
}

func (b *Bot) onMessage(message tgbotapi.Message, bot_api *tgbotapi.BotAPI) {
	var resp string
	switch {
	case strings.HasPrefix(strings.ToLower(message.Text), "!пинг"):
		resp = ping(message)
	case strings.HasPrefix(strings.ToLower(message.Text), "!время"):
		resp = getUsersCurrentTime(b.DataDir)
	case strings.HasPrefix(strings.ToLower(message.Text), "!погода"):
		resp = getCurrentWeather(b.WeatherAPIKey, b.WeatherAPICities)
	case strings.HasPrefix(strings.ToLower(message.Text), "!чат"):
		resp = askChatGPT(b.OpenaiAPIKey, strings.TrimPrefix(message.Text, "!чат"))
	case strings.HasPrefix(strings.ToLower(message.Text), "!команды"):
		resp = "!пинг - проверка связи\n!время - текущее время у участников чата\n!погода - текущая погода\n!чат - вопрос к ChatGPT\n!команды - список доступных команд"
	case message.Location != nil && message.From != nil:
		rememberTZ(message, b.DataDir)
		return
	default:
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, resp)

	if _, err := bot_api.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (b *Bot) mourningJob(bot_api *tgbotapi.BotAPI) {
	log.Println("[INFO] Starting mourning job")
	for {
		text := "Доброе утро! 🌅\n"
		waitUntilMourning()
		// call currency api
		c, err := apiclient.GetCurrency(b.CurrencyAPIKey)
		if err != nil {
			log.Printf("[ERROR] Could not get currency exchange rates: %v", err)
		} else {
			text += fmt.Sprintf("\nКурсы валют:\n    USD %.2f₽\n    EUR %.2f₽\n    BTC %.2f$\n", c.Data.RUB.Value, c.Data.RUB.Value/c.Data.EUR.Value, 1.0/c.Data.BTC.Value)
		}
		// call weather api
		weather := apiclient.Get_weather(b.WeatherAPIKey, b.WeatherAPICities)
		sort.Slice(weather, func(i, j int) bool {
			return weather[i].CurrentWeather.Temp < weather[j].CurrentWeather.Temp
		})
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

func waitUntilMourning() {
	t := time.Now()
	desiredTime := time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())
	if desiredTime.Sub(t) <= 0 {
		desiredTime = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location()).Add(24 * time.Hour)
	}
	log.Println("[INFO] Waiting until mourning ", desiredTime.Sub(t))
	time.Sleep(desiredTime.Sub(t))
}
