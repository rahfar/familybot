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
	Token         string
	Dbg           bool
	Chats         []string
	GroupID       int64
	DataDir       string
	TGBotAPI      *tgbotapi.BotAPI
	AnekdotAPI    *apiclient.AnecdoteAPI
	ExchangeAPI   *apiclient.ExchangeAPI
	SheetsAPI     *apiclient.SheetsAPI
	KommersantAPI *apiclient.KommersantAPI
	OpenaiAPI     *apiclient.OpenaiAPI
	WeatherAPI    *apiclient.WeatherAPI
}

func (b *Bot) Run() {
	usernames := make(map[string]struct{}, 0)
	for _, username := range b.Chats {
		usernames[username] = struct{}{}
	}

	go b.mourningJob()

	update_cfg := tgbotapi.NewUpdate(0)
	update_cfg.Timeout = 60

	updates := b.TGBotAPI.GetUpdatesChan(update_cfg)

	for update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}
		_, ok := usernames[update.Message.Chat.UserName]
		ok = ok || update.Message.Chat.ID == b.GroupID
		if !ok {
			log.Printf("[INFO] Skip message from unsupported chat. Chat: %+v\n", *update.Message.Chat)
			continue
		}
		go b.onMessage(*update.Message)
	}
}

func (b *Bot) onMessage(message tgbotapi.Message) {
	var resp string
	var pm string
	var disable_web_page_preview bool
	switch {
	case strings.HasPrefix(strings.ToLower(message.Text), "!пинг"):
		resp = ping(message)
	case strings.HasPrefix(strings.ToLower(message.Text), "!погода"):
		resp = getCurrentWeather(b.WeatherAPI)
	case strings.HasPrefix(strings.ToLower(message.Text), "!чат"):
		resp = askChatGPT(b.OpenaiAPI, strings.TrimPrefix(message.Text, "!чат"))
	case strings.HasPrefix(strings.ToLower(message.Text), "!продажи"):
		resp = getYesterdaySales(b.SheetsAPI)
	case strings.HasPrefix(strings.ToLower(message.Text), "!анекдот"):
		resp = getAnecdote(b.AnekdotAPI)
	case strings.HasPrefix(strings.ToLower(message.Text), "!новости"):
		resp = getLatestNews(b.KommersantAPI)
		pm = tgbotapi.ModeMarkdown
		disable_web_page_preview = true
	case strings.HasPrefix(strings.ToLower(message.Text), "!команды"):
		resp = "!пинг - проверка связи\n!погода - текущая погода\n!чат - вопрос к ChatGPT\n!команды - список доступных команд\n!продажи - текущие продажи из google spreadsheet\n!анекдот - случайный анекдот\n!новости - последние 3 новости из Коммерсанта"
	case message.Voice != nil:
		resp = transcriptVoice(b.OpenaiAPI, b.TGBotAPI, message.Voice.FileID)
	default:
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, resp)
	msg.ParseMode = pm
	msg.DisableWebPagePreview = disable_web_page_preview
	msg.ReplyToMessageID = message.MessageID
	if _, err := b.TGBotAPI.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (b *Bot) mourningJob() {
	log.Println("[INFO] Starting mourning job")
	for {
		text := "Доброе утро! 🌅\n"
		waitUntilMourning()
		// call currency api
		xr_today, err1 := b.ExchangeAPI.GetExchangeRates()
		xr_yesterday, err2 := b.ExchangeAPI.GetHistoryExchangeRates(time.Now().UTC().Add(-48 * time.Hour))
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

			text += fmt.Sprintf("\nКурсы валют:\n    USD %.2f₽ (%+.2f%%) \n    EUR %.2f₽ (%+.2f%%)\n    BTC %.2f$ (%+.2f%%)\n",
				USDRUB_today,
				(USDRUB_today/USDRUB_yesterday-1)*100,
				EURRUB_today,
				(EURRUB_today/EURRUB_yesterday-1)*100,
				BTCUSD_today,
				(BTCUSD_today/BTCUSD_yesterday-1)*100,
			)
		}
		// call weather api
		weather := b.WeatherAPI.GetWeather()
		sort.Slice(weather, func(i, j int) bool {
			return weather[i].Current.Temp < weather[j].Current.Temp
		})
		if len(weather) > 0 {
			text += "\nПрогноз погоды:\n"
			for _, w := range weather {
				text += fmt.Sprintf("    %s: %+g°C (max: %+g°C, min: %+g°C), %s \n", w.Location.Name, w.Current.Temp, w.Forecast.Forecastday[0].Day.Maxtemp_c, w.Forecast.Forecastday[0].Day.Mintemp_c, w.Current.Condition.Text)
			}
		}
		//call news api
		text += getLatestNews(b.KommersantAPI)
		// send message to group
		msg := tgbotapi.NewMessage(b.GroupID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = true
		if _, err := b.TGBotAPI.Send(msg); err != nil {
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
