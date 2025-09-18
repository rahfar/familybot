package bot

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rahfar/familybot/src/apiclient"
	"github.com/rahfar/familybot/src/db"
	"github.com/rahfar/familybot/src/metrics"
)

type Bot struct {
	Token        string
	Dbg          bool
	Host         string
	Port         string
	AdminUserIDs []int64
	GroupID      int64
	AskGPTCache  *expirable.LRU[string, []apiclient.GPTResponse]
	TGBotAPI     *tgbotapi.BotAPI
	ExchangeAPI  *apiclient.ExchangeAPI
	OpenaiAPI    *apiclient.OpenaiAPI
	WeatherAPI   *apiclient.WeatherAPI
	MinifluxAPI  *apiclient.MinifluxAPI
	DeeplAPI     *apiclient.DeeplAPI
	DBClient     *db.Client
}

func (b *Bot) Run() {
	go b.startWebAPI()
	go b.mourningJob()

	_, err := b.initCommands()
	if err != nil {
		slog.Error("couldn't init commnads", "err", err)
		panic(err)
	}

	update_cfg := tgbotapi.NewUpdate(0)
	update_cfg.Timeout = 60
	updates := b.TGBotAPI.GetUpdatesChan(update_cfg)
	for update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "/start") || strings.HasPrefix(update.Message.Text, "/whoami") {
			// Always allow /start and /whoami commands
			slog.Info("received /start or /whoami command", "from", update.Message.From, "chat", update.Message.Chat)
		} else {
			// Check if the chat is authorized
			if !b.isChatAuthorized(*update.Message) {
				slog.Info("skip message from unsupported chat", "chat", *update.Message.Chat)
				if update.Message.Chat.IsPrivate() {
					unauthorizedResponse := fmt.Sprintf(
						"У вас нет прав на общение с этим ботом. Пожалуйста, свяжитесь с администратором. "+
							"Chat ID: %d",
						update.Message.Chat.ID,
					)
					msgConfig := tgbotapi.NewMessage(update.Message.Chat.ID, unauthorizedResponse)
					msgConfig.ReplyToMessageID = update.Message.MessageID
					b.sendMessage(msgConfig)
				}
				return
			}
		}

		metrics.RecvMsgCounter.Inc()

		go b.onMessage(*update.Message)
	}
}

func findCommand(msgText string) *Command {
	for _, cmd := range Commands {
		if strings.HasPrefix(msgText, cmd.Name) {
			return &cmd
		}
	}
	return nil
}

func (b *Bot) onMessage(msg tgbotapi.Message) {
	slog.Debug("received message", "message", msg)

	cmd := findCommand(msg.Text)

	if cmd != nil {
		slog.Debug("command found", "command", cmd.Name)
		metrics.CommandCallsCaounter.With(prometheus.Labels{"command": cmd.Name}).Inc()
		cmd.Handler(b, &msg)
	} else if msg.Voice != nil {
		transcriptVoice(b, &msg)
	} else if msg.Chat.IsPrivate() {
		cmd, exists := Commands["/gpt"]
		if !exists {
			slog.Error("could not find command /gpt")
			return
		}
		cmd.Handler(b, &msg)
	} else {
		slog.Info("unsupported command")
		return
	}
}

func (b *Bot) mourningDigest() string {
	text := "Доброе утро\\! 🌅\n"

	// call currency api
	xr_today, err1 := b.ExchangeAPI.GetExchangeRates(time.Now().UTC())
	xr_yesterday, err2 := b.ExchangeAPI.GetExchangeRates(time.Now().UTC().Add(-48 * time.Hour))
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

		text += "\n_Курсы валют:_\n" + tgbotapi.EscapeText(
			tgbotapi.ModeMarkdownV2,
			fmt.Sprintf(
				"USD %.2f₽ (%+.2f%%) \nEUR %.2f₽ (%+.2f%%)\nBTC %.2f$ (%+.2f%%)\n",
				USDRUB_today,
				(USDRUB_today/USDRUB_yesterday-1)*100,
				EURRUB_today,
				(EURRUB_today/EURRUB_yesterday-1)*100,
				BTCUSD_today,
				(BTCUSD_today/BTCUSD_yesterday-1)*100,
			),
		)
	}

	// call weather api
	weather := b.WeatherAPI.GetWeather()
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].List[0].Main.Temp < weather[j].List[0].Main.Temp
	})
	if len(weather) > 0 {
		text += "\n_Прогноз погоды:_\n"
		for _, w := range weather {
			location := time.FixedZone("custom", w.City.Timezone)
			sunriseTime := time.Unix(w.City.Sunrise, 0).In(location).Format("15:04")
			sunsetTime := time.Unix(w.City.Sunset, 0).In(location).Format("15:04")
			minTemp, maxTemp := b.WeatherAPI.GetMinMaxTemp(w)
			text += fmt.Sprintf("*%s:*\n", w.City.Name)
			text += tgbotapi.EscapeText(
				tgbotapi.ModeMarkdownV2,
				fmt.Sprintf(
					"  %d°C (min: %d°C, max: %d°C), %s\n  восход: %s закат: %s\n",
					int(w.List[0].Main.Temp),
					int(minTemp),
					int(maxTemp),
					w.List[0].Weather[0].Description,
					sunriseTime,
					sunsetTime,
				),
			)
		}
	}

	nt_news, nt_err := b.MinifluxAPI.GetLatestNews("https://www.nytimes.com", 3)
	tass_news, tass_err := b.MinifluxAPI.GetLatestNews("https://tass.ru", 2)
	if (nt_err != nil) || (len(nt_news) == 0) || (tass_err != nil) || (len(tass_news) == 0) {
		slog.Error("error calling news api", "err", nt_err, "tass_err", tass_err)
	} else {
		fmt_news := "\n_Последние новости:_\nNew York Times\n"
		i := 1
		for _, n := range nt_news {
			translatedTitle, err := b.DeeplAPI.CallDeeplAPI([]string{n.Title})
			if err != nil {
				slog.Error("error calling deepl api", "err", err)
				translatedTitle = n.Title
			}
			translatedTitle = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, translatedTitle)
			escaped_url := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, n.URL)
			fmt_news += fmt.Sprintf("%d\\. [%s](%s)\n", i, translatedTitle, escaped_url)
			i++
		}
		fmt_news += "ТАСС\n"
		for _, n := range tass_news {
			title := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, n.Title)
			escaped_url := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, n.URL)
			fmt_news += fmt.Sprintf("%d\\. [%s](%s)\n", i, title, escaped_url)
			i++
		}
		text += fmt_news
	}
	return text
}

func (b *Bot) mourningJob() {
	metrics.MourningJobCounter.Inc()
	slog.Info("starting mourning job")
	for {
		waitUntilMourning()
		text := b.mourningDigest()
		// send message to group
		msg := tgbotapi.NewMessage(b.GroupID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.DisableWebPagePreview = true

		b.sendMessage(msg)
	}
}

func (b *Bot) isChatAuthorized(msg tgbotapi.Message) bool {
	chatID := msg.Chat.ID

	// Always allow the main group
	if chatID == b.GroupID {
		return true
	}

	// For private chats, check if the user is admin
	if msg.Chat.IsPrivate() && b.isUserAdmin(msg.From.ID) {
		return true
	}

	// Check if this chat (private or group) is authorized
	ctx := context.Background()
	authorized, err := b.DBClient.IsChatAuthorized(ctx, chatID)
	if err != nil {
		slog.Error("error checking chat authorization", "err", err, "chat_id", chatID)
		return false
	}

	return authorized
}

func (b *Bot) isUserAdmin(userID int64) bool {
	return slices.Contains(b.AdminUserIDs, userID)
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

	for i := range msgParts {
		start := i * maxMsgLength
		end := min((i+1)*maxMsgLength, msgLength)

		msg.Text = strings.ToValidUTF8(msgText[start:end], "")

		for i := 1; i <= maxRetry; i++ {
			_, err := b.TGBotAPI.Send(msg)
			if err == nil {
				metrics.SentMsgCounter.Inc()
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

func waitUntilMourning() {
	t := time.Now()
	desiredTime := time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())
	if desiredTime.Sub(t) <= 5*time.Second {
		desiredTime = desiredTime.Add(24 * time.Hour)
	}
	slog.Info("waiting until mourning", "time-to-wait", desiredTime.Sub(t).String())
	time.Sleep(desiredTime.Sub(t))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (b *Bot) startWebAPI() {
	http.HandleFunc("/ping", pingHandler)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(b.Host+":"+b.Port, nil)
	panic(err)
}

func (b *Bot) initCommands() (*tgbotapi.APIResponse, error) {
	tgCommands := make([]tgbotapi.BotCommand, 0, len(Commands))
	AddListCommand()
	for _, cmd := range Commands {
		if cmd.Hidden {
			continue
		}
		tgCommands = append(tgCommands, tgbotapi.BotCommand{
			Command:     cmd.Name,
			Description: cmd.Description,
		})
	}

	cmdCfg := tgbotapi.NewSetMyCommands(tgCommands...)
	return b.TGBotAPI.Request(cmdCfg)
}
