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
	Token                        string
	Dbg                          bool
	GroupID                      int64
	AdminChatID                  int64
	DataDir                      string
	WeatherAPIKey                string
	WeatherAPICities             []string
	CurrencyAPIKey               string
	OpenaiAPIKey                 string
	GoogleSheetsAPIKey           string
	GoogleSheetsAPISpreadSheetID string
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
		go b.onMessage(*update.Message, bot_api)
	}
}

func (b *Bot) onMessage(message tgbotapi.Message, bot_api *tgbotapi.BotAPI) {
	var resp string
	var pm string
	switch {
	case strings.HasPrefix(strings.ToLower(message.Text), "!пинг"):
		resp = ping(message)
	case strings.HasPrefix(strings.ToLower(message.Text), "!время"):
		resp = getUsersCurrentTime(b.DataDir)
	case strings.HasPrefix(strings.ToLower(message.Text), "!погода"):
		resp = getCurrentWeather(b.WeatherAPIKey, b.WeatherAPICities)
	case strings.HasPrefix(strings.ToLower(message.Text), "!чат"):
		resp = askChatGPT(b.OpenaiAPIKey, strings.TrimPrefix(message.Text, "!чат"))
	case strings.HasPrefix(strings.ToLower(message.Text), "!продажи"):
		resp = getYesterdaySales(b.GoogleSheetsAPIKey, b.GoogleSheetsAPISpreadSheetID)
	case strings.HasPrefix(strings.ToLower(message.Text), "!анекдот"):
		resp = getAnecdote()
	case strings.HasPrefix(strings.ToLower(message.Text), "!новости"):
		resp = getLatestNews()
		pm = tgbotapi.ModeMarkdown
	case strings.HasPrefix(strings.ToLower(message.Text), "!команды"):
		resp = "!пинг - проверка связи\n!время - текущее время у участников чата\n!погода - текущая погода\n!чат - вопрос к ChatGPT\n!команды - список доступных команд\n!продажи - текущие продажи из google spreadsheet\n!анекдот - случайный анекдот\n!новости - последние 3 новости из Коммерсанта"
	case message.Location != nil && message.From != nil:
		rememberTZ(message, b.DataDir)
		return
	default:
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, resp)
	msg.ParseMode = pm
	msg.ReplyToMessageID = message.MessageID
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
		xr_today, err1 := apiclient.GetExchangeRates(b.CurrencyAPIKey)
		xr_yesterday, err2 := apiclient.GetHistoryExchangeRates(b.CurrencyAPIKey, time.Now().UTC().Add(-48*time.Hour))
		switch {
		case err1 != nil:
			log.Printf("[ERROR] Could not get currency exchange rates: %v", err1)
		case err2 != nil:
			log.Printf("[ERROR] Could not get currency history exchange rates: %v", err2)
		default:
			USDRUB_today := xr_today.Data.RUB.Value
			EURRUB_today := xr_today.Data.RUB.Value / xr_today.Data.EUR.Value
			BTCUSD_today := 1.0 / xr_today.Data.BTC.Value
			USDRUB_yesterday := xr_yesterday.Data.RUB.Value
			EURRUB_yesterday := xr_yesterday.Data.RUB.Value / xr_yesterday.Data.EUR.Value
			BTCUSD_yesterday := 1.0 / xr_yesterday.Data.BTC.Value

			text += fmt.Sprintf("\nКурсы валют:\n    USD %.2f₽ (%.2f%%) \n    EUR %.2f₽ (%.2f%%)\n    BTC %.2f$ (%.2f%%)\n",
				USDRUB_today,
				(USDRUB_today/USDRUB_yesterday-1)*100,
				EURRUB_today,
				(EURRUB_today/EURRUB_yesterday-1)*100,
				BTCUSD_today,
				(BTCUSD_today/BTCUSD_yesterday-1)*100,
			)
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
		//call news api
		text += getLatestNews()
		// send message to group
		msg := tgbotapi.NewMessage(b.GroupID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err := bot_api.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func waitUntilMourning() {
	t := time.Now()
	desiredTime := time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())
	if desiredTime.Sub(t) <= 5*time.Second {
		desiredTime = desiredTime.Add(24 * time.Hour)
	}
	log.Println("[INFO] Waiting until mourning ", desiredTime.Sub(t))
	time.Sleep(desiredTime.Sub(t))
}
