package bot

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"

	"github.com/rahfar/familybot/src/apiclient"
)

func ping(b *Bot, msg *tgbotapi.Message) {
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "понг")
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func getCurrentWeather(b *Bot, msg *tgbotapi.Message) {
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "")
	weather := b.WeatherAPI.GetWeather()
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].Current.Temp < weather[j].Current.Temp
	})
	if len(weather) > 0 {
		for _, w := range weather {
			msgConfig.Text += fmt.Sprintf("%s: %+g°C (max: %+g°C, min: %+g°C), %s\n", w.Location.Name, w.Current.Temp, w.Forecast.Forecastday[0].Day.Maxtemp_c, w.Forecast.Forecastday[0].Day.Mintemp_c, w.Current.Condition.Text)
		}
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	} else {
		msgConfig = tgbotapi.NewMessage(msg.Chat.ID, "Нет данных")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
}

func askChatGPT(b *Bot, msg *tgbotapi.Message) {
	question := removeFirstWord(msg.Text)

	responseHistory, ok := b.AskGPTCache.Get(strconv.FormatInt(msg.Chat.ID, 10))
	if !ok {
		responseHistory = make([]apiclient.GPTResponse, 0)
	}
	responseHistory = filterOldGPTResponce(responseHistory)

	ans, err := b.OpenaiAPI.CallGPT3dot5(question, responseHistory)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	responseHistory = append(responseHistory, apiclient.GPTResponse{Role: openai.ChatMessageRoleAssistant, Content: ans, Time: time.Now()})
	responseHistory = append(responseHistory, apiclient.GPTResponse{Role: openai.ChatMessageRoleUser, Content: question, Time: time.Now()})
	b.AskGPTCache.Add(strconv.FormatInt(msg.Chat.ID, 10), responseHistory)

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, ans)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.DisableWebPagePreview = true
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func filterOldGPTResponce(responseHistory []apiclient.GPTResponse) []apiclient.GPTResponse {
	filtered := make([]apiclient.GPTResponse, 0)
	for _, v := range responseHistory {
		if v.Time.After(time.Now().Add(-5 * time.Minute)) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func getAnecdote(b *Bot, msg *tgbotapi.Message) {
	anecdote, err := b.AnekdotAPI.CallAnecdoteAPI()
	if err != nil || len(anecdote) == 0 {
		slog.Error("error calling anecdote api", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Не смог получить свежий анекдот :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, anecdote)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func getLatestNews(b *Bot, msg *tgbotapi.Message) {
	news, err := b.KommersantAPI.CallKommersantAPI()
	if (err != nil) || (len(news) == 0) {
		slog.Error("error calling news api", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Не смог получить последние новости :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	fmt_news := "\nПоследние новости:\n"
	for i, n := range news[:3] {
		fmt_news += fmt.Sprintf("%d. [%s](%s)\n", i+1, n.Title, n.Link)
	}
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, fmt_news)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.DisableWebPagePreview = true
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func transcriptVoice(b *Bot, msg *tgbotapi.Message) {
	// Get direct link to audio message
	link, err := b.TGBotAPI.GetFileDirectURL(msg.Voice.FileID)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	filename := "/tmp/" + msg.Voice.FileID + path.Ext(link)
	slog.Debug("saving audio", "filename", filename)
	// Download audio file
	resp, err := http.Get(link)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	defer resp.Body.Close()

	// Create the output file
	file, err := os.Create(filename)
	if err != nil {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	defer file.Close()
	defer os.Remove(filename)

	// Write the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	text, err := b.OpenaiAPI.CallWhisper(filename)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обработки голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func generateImage(b *Bot, msg *tgbotapi.Message) {
	prompt := removeFirstWord(msg.Text)
	imgURL, err := b.OpenaiAPI.CallDalle(prompt)
	if err != nil {
		slog.Error("generating image", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при генерации картинки :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	photoConfig := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(imgURL))
	photoConfig.ReplyToMessageID = msg.MessageID
	b.sendPhoto(photoConfig)
}

func removeFirstWord(input string) string {
	// Find the index of the first space
	firstSpaceIndex := strings.Index(input, " ")

	// If no space is found, return the original string
	if firstSpaceIndex == -1 {
		return input
	}

	// Extract the substring after the first space
	// (adding 1 to exclude the space itself)
	result := input[firstSpaceIndex+1:]

	return result
}
