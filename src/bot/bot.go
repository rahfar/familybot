package bot

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/rahfar/familybot/src/apiclient"
)

type Bot struct {
	Token            string
	Dbg              bool
	AllowedUsernames []string
	GroupID          int64
	Commands         []Command
	AskGPTCache      *expirable.LRU[string, []apiclient.GPTResponse]
	TGBotAPI         *tgbotapi.BotAPI
	AnekdotAPI       *apiclient.AnecdoteAPI
	ExchangeAPI      *apiclient.ExchangeAPI
	KommersantAPI    *apiclient.KommersantAPI
	OpenaiAPI        *apiclient.OpenaiAPI
	WeatherAPI       *apiclient.WeatherAPI
	MinifluxAPI      *apiclient.MinifluxAPI
	DeeplAPI         *apiclient.DeeplAPI
}

func (b *Bot) Run() {
	go b.mourningJob()

	update_cfg := tgbotapi.NewUpdate(0)
	update_cfg.Timeout = 60

	updates := b.TGBotAPI.GetUpdatesChan(update_cfg)

	for update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}

		if !b.isMessageFromAllowedChat(update) {
			slog.Info("skip message from unsupported chat", "chat", *update.Message.Chat)
			continue
		}
		go b.onMessage(*update.Message)
	}
}

func (b *Bot) onMessage(msg tgbotapi.Message) {
	words := strings.Split(msg.Text, " ")
	cmd := findCommand(b.Commands, words[0])
	if cmd != nil {
		cmd.Handler(b, &msg)
	} else if strings.EqualFold(words[0], "!команды") {
		help_text := "Доступные команды:\n"
		for _, c := range b.Commands {
			help_text += strings.Join(c.Names, ", ") + " - " + c.Description + "\n"
		}
		resp := tgbotapi.NewMessage(msg.Chat.ID, help_text)
		resp.ReplyToMessageID = msg.MessageID
		b.sendMessage(resp)
	} else if msg.Voice != nil {
		transcriptVoice(b, &msg)
	} else {
		return
	}
}

func (b *Bot) mourningJob() {
	slog.Info("starting mourning job")
	for {
		text := "Доброе утро! 🌅\n"
		waitUntilMourning()

		// call currency api
		xr_today, err1 := b.ExchangeAPI.GetExchangeRates()
		xr_yesterday, err2 := b.ExchangeAPI.GetHistoryExchangeRates(time.Now().UTC().Add(-48 * time.Hour))
		switch {
		case err1 != nil:
			slog.Error("could not get currency exchange rates", "err", err1)
		case err2 != nil:
			slog.Error("could not get currency history exchange rates", "err", err2)
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

		// call news api
		news, err := b.MinifluxAPI.GetLatestNews()
		if (err != nil) || (len(news) == 0) {
			slog.Error("error calling news api", "err", err)
		} else {
			fmt_news := "\nПоследние новости:\n"
			for i, n := range news {
				translatedTitle, err := b.DeeplAPI.CallDeeplAPI(n.Title)
				if err != nil {
					slog.Error("error calling deepl api", "err", err)
					translatedTitle = n.Title
				}
				fmt_news += fmt.Sprintf("%d. [%s](%s)\n", i+1, translatedTitle, n.URL)
			}
			text += fmt_news
		}

		// send message to group
		msg := tgbotapi.NewMessage(b.GroupID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = true

		b.sendMessage(msg)
	}
}

func (b *Bot) isMessageFromAllowedChat(update tgbotapi.Update) bool {
	if update.Message.Chat.ID == b.GroupID {
		return true
	}
	for _, un := range b.AllowedUsernames {
		if un == update.Message.Chat.UserName {
			return true
		}
	}
	return false
}

func (b *Bot) sendMessage(msg tgbotapi.MessageConfig) {
	const (
		maxRetry     = 3
		maxMsgLength = 4096
	)

	msgText := msg.Text
	msgLength := len(msgText)

	if msgLength == 0 {
		slog.Info("zero length msg would not be send")
		return
	}

	msgParts := (msgLength + maxMsgLength - 1) / maxMsgLength // Ceiling division

	for i := 0; i < msgParts; i++ {
		start := i * maxMsgLength
		end := (i + 1) * maxMsgLength
		if end > msgLength {
			end = msgLength
		}

		msg.Text = strings.ToValidUTF8(msgText[start:end], "")

		for i := 1; i <= maxRetry; i++ {
			_, err := b.TGBotAPI.Send(msg)
			if err == nil {
				break
			}

			if i < maxRetry {
				slog.Info(
					"error sending message, retrying in 5 seconds (disable formatting)...",
					"err", err,
					"message", msg.Text,
					"retry-cnt", i,
				)
				msg.ParseMode = ""
				time.Sleep(5 * time.Second)
			} else {
				slog.Error("error sending message", "err", err, "message", msg.Text)
				return
			}
		}
	}
}

func (b *Bot) sendPhoto(msg tgbotapi.PhotoConfig) {
	const maxRetry = 3

	for i := 1; i <= maxRetry; i++ {
		_, err := b.TGBotAPI.Send(msg)
		if err == nil {
			break
		}

		if i < maxRetry {
			slog.Info("error sending photo, retrying in 5 seconds...", "err", err, "retry-cnt", i)
			time.Sleep(5 * time.Second)
		} else {
			slog.Error("error sending photo", "err", err)
			return
		}
	}
}

func waitUntilMourning() {
	t := time.Now()
	desiredTime := time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())
	if desiredTime.Sub(t) <= 5*time.Second {
		desiredTime = desiredTime.Add(24 * time.Hour)
	}
	slog.Info("waiting until mourning", "time-to-wait", desiredTime.Sub(t).String())
	time.Sleep(desiredTime.Sub(t))
}
