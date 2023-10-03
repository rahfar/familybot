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

func ping(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(msg.Chat.ID, "понг")
}

func getCurrentWeather(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	resp := tgbotapi.NewMessage(msg.Chat.ID, "")
	weather := bot.WeatherAPI.GetWeather()
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].Current.Temp < weather[j].Current.Temp
	})
	if len(weather) > 0 {
		for _, w := range weather {
			resp.Text += fmt.Sprintf("%s: %+g°C (max: %+g°C, min: %+g°C), %s\n", w.Location.Name, w.Current.Temp, w.Forecast.Forecastday[0].Day.Maxtemp_c, w.Forecast.Forecastday[0].Day.Mintemp_c, w.Current.Condition.Text)
		}
		return resp
	} else {
		return tgbotapi.NewMessage(msg.Chat.ID, "Нет данных")
	}
}

func askChatGPT(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	question := removeFirstWord(msg.Text)

	responseHistory, ok := bot.AskGPTCache.Get(strconv.FormatInt(msg.Chat.ID, 10))
	if !ok {
		responseHistory = make([]apiclient.GPTResponse, 0)
	}
	responseHistory = filterOldGPTResponce(responseHistory)

	ans, err := bot.OpenaiAPI.CallGPT3dot5(question, responseHistory)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
	}

	responseHistory = append(responseHistory, apiclient.GPTResponse{Role: openai.ChatMessageRoleAssistant, Content: ans, Time: time.Now()})
	responseHistory = append(responseHistory, apiclient.GPTResponse{Role: openai.ChatMessageRoleUser, Content: question, Time: time.Now()})
	bot.AskGPTCache.Add(strconv.FormatInt(msg.Chat.ID, 10), responseHistory)

	resp := tgbotapi.NewMessage(msg.Chat.ID, ans)
	resp.ParseMode = tgbotapi.ModeMarkdown
	resp.DisableWebPagePreview = true
	return resp
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

func getAnecdote(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	anecdote, err := bot.AnekdotAPI.CallAnecdoteAPI()
	if err != nil || len(anecdote) == 0 {
		slog.Error("error calling anecdote api", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Не смог получить свежий анекдот :(")
	}
	return tgbotapi.NewMessage(msg.Chat.ID, anecdote)
}

func getLatestNews(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	news, err := bot.KommersantAPI.CallKommersantAPI()
	if (err != nil) || (len(news) == 0) {
		slog.Error("error calling news api", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Не смог получить последние новости :(")
	}
	fmt_news := "\nПоследние новости:\n"
	for i, n := range news[:3] {
		fmt_news += fmt.Sprintf("%d. [%s](%s)\n", i+1, n.Title, n.Link)
	}
	resp := tgbotapi.NewMessage(msg.Chat.ID, fmt_news)
	resp.ParseMode = tgbotapi.ModeMarkdown
	resp.DisableWebPagePreview = true
	return resp
}

func transcriptVoice(bot *Bot, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	// Get direct link to audio message
	link, err := bot.TGBotAPI.GetFileDirectURL(msg.Voice.FileID)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")

	}
	filename := "/tmp/" + msg.Voice.FileID + path.Ext(link)
	slog.Debug("saving audio", "filename", filename)
	// Download audio file
	resp, err := http.Get(link)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
	}
	defer resp.Body.Close()

	// Create the output file
	file, err := os.Create(filename)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
	}
	defer file.Close()
	defer os.Remove(filename)

	// Write the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при скачивании голосового сообщения")
	}

	text, err := bot.OpenaiAPI.CallWhisper(filename)
	if err != nil {
		slog.Error("getting voice msg", "err", err)
		return tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обработки голосового сообщения")
	}
	return tgbotapi.NewMessage(msg.Chat.ID, text)
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
