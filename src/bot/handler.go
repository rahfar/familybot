package bot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
		return weather[i].List[0].Main.Temp < weather[j].List[0].Main.Temp
	})
	if len(weather) > 0 {
		for _, w := range weather {
			location := time.FixedZone("custom", w.City.Timezone)
			sunriseTime := time.Unix(w.City.Sunrise, 0).In(location).Format("15:04")
			sunsetTime := time.Unix(w.City.Sunset, 0).In(location).Format("15:04")
			minTemp, maxTemp := b.WeatherAPI.GetMinMaxTemp(w)
			msgConfig.Text += fmt.Sprintf("*%s:*\n", w.City.Name)
			msgConfig.Text += tgbotapi.EscapeText(
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
		msgConfig.ReplyToMessageID = msg.MessageID
		msgConfig.ParseMode = tgbotapi.ModeMarkdownV2
		b.sendMessage(msgConfig)
		return
	} else {
		msgConfig = tgbotapi.NewMessage(msg.Chat.ID, "Нет данных")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
}

func getRevision(b *Bot, msg *tgbotapi.Message) {
	rev := os.Getenv("REVISION")
	if len(rev) == 0 {
		return
	}
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, rev)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func whoAmI(b *Bot, msg *tgbotapi.Message) {
	if !msg.Chat.IsPrivate() && !msg.Chat.IsGroup() {
		return
	}
	chatId := msg.Chat.ID
	userId := msg.From.ID
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("ChatID: %d\nUserID: %d", chatId, userId))
	b.sendMessage(msgConfig)
}

func sendMourningDigest(b *Bot, msg *tgbotapi.Message) {
	text := b.mourningDigest()
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdownV2
	msgConfig.DisableWebPagePreview = true
	b.sendMessage(msgConfig)
}

func askChatGPT(b *Bot, msg *tgbotapi.Message) {
	var question string

	if msg.IsCommand() {
		question = strings.TrimSpace(msg.CommandArguments())
	} else if strings.HasPrefix(msg.Text, "/gpt@") {
		words := strings.SplitN(msg.Text, " ", 2)
		if len(words) == 2 {
			question = strings.TrimSpace(words[1])
		}

	} else if strings.HasPrefix(msg.Text, "/gpt") {
		question = strings.TrimSpace(msg.Text[len("/gpt"):])
	} else {
		question = strings.TrimSpace(msg.Text)
	}

	if len(question) == 0 {
		slog.Debug("empty question")
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Пустой входной вопрос")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	slog.Debug("askChatGPT", "question", question)

	responseHistory, ok := b.AskGPTCache.Get(strconv.FormatInt(msg.Chat.ID, 10))
	if !ok {
		responseHistory = make([]apiclient.GPTResponse, 0)
	}
	// DISABLED auto cleanup old messages
	// responseHistory = filterOldGPTResponce(responseHistory)

	ans, err := b.OpenaiAPI.GenerateChatCompletion(question, responseHistory)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	responseHistory = append(responseHistory, apiclient.GPTResponse{
		Role:    openai.ChatMessageRoleAssistant,
		Content: ans,
		Time:    time.Now(),
	})
	responseHistory = append(responseHistory, apiclient.GPTResponse{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
		Time:    time.Now(),
	})
	b.AskGPTCache.Add(strconv.FormatInt(msg.Chat.ID, 10), responseHistory)

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, ans)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.DisableWebPagePreview = true
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func filterOldGPTResponce(responseHistory []apiclient.GPTResponse) []apiclient.GPTResponse {
	const DurationChatHistory = -24 * 60 * time.Minute

	filtered := make([]apiclient.GPTResponse, 0)
	for _, v := range responseHistory {
		if v.Time.After(time.Now().Add(DurationChatHistory)) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func newChatGPT(b *Bot, msg *tgbotapi.Message) {
	b.AskGPTCache.Remove(strconv.FormatInt(msg.Chat.ID, 10))
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Контекст вызова GPT удален")
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

	// Convert the audio file to mp3
	mp3Filename := filename + ".mp3"
	err = convertOgaToMp3(filename, mp3Filename)
	if err != nil {
		slog.Error("converting voice msg", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при обработки голосового сообщения")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	defer os.Remove(mp3Filename)

	text, err := b.OpenaiAPI.TranscribeAudioFile(mp3Filename)
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

func correctEnglish(b *Bot, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.CommandArguments())
	if len(text) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Пустой входной вопрос")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	ans, err := b.OpenaiAPI.CorrectGrammarAndStyle(text)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, ans)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func translateEng2Ru(b *Bot, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.CommandArguments())
	if len(text) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Пустой входной вопрос")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	ans, err := b.OpenaiAPI.TranslateEnglishToRussian(text)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, ans)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func translateRu2Eng(b *Bot, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.CommandArguments())
	if len(text) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Пустой входной вопрос")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	ans, err := b.OpenaiAPI.TranslateRussianToEnglish(text)
	if err != nil || len(ans) == 0 {
		slog.Error("error occured while call openai", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при вызове ChatGPT :(")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, ans)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func listCommands(b *Bot, msg *tgbotapi.Message) {
	var text string
	for _, cmd := range Commands {
		text += fmt.Sprintf("%s - %s\n", cmd.Name, cmd.Description)
	}
	if len(text) == 0 {
		text = "Нет доступных команд."
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func addUser(b *Bot, msg *tgbotapi.Message) {
	if !b.isUserAdmin(msg.From.ID) {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	chatIDStr := strings.TrimSpace(msg.CommandArguments())
	if len(chatIDStr) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Укажите chat ID для добавления")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Неверный формат chat ID")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	ctx := context.Background()
	err = b.DBClient.AddChat(ctx, chatID)
	if err != nil {
		slog.Error("error adding chat", "err", err, "chat_id", chatID)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при добавлении чата")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Чат %d добавлен в список авторизованных", chatID))
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func removeUser(b *Bot, msg *tgbotapi.Message) {
	if !b.isUserAdmin(msg.From.ID) {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	chatIDStr := strings.TrimSpace(msg.CommandArguments())
	if len(chatIDStr) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Укажите chat ID для удаления")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Неверный формат chat ID")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	ctx := context.Background()
	err = b.DBClient.RemoveChat(ctx, chatID)
	if err != nil {
		slog.Error("error removing chat", "err", err, "chat_id", chatID)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при удалении чата")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Чат %d удален из списка авторизованных", chatID))
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func listUsers(b *Bot, msg *tgbotapi.Message) {
	if !b.isUserAdmin(msg.From.ID) {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	ctx := context.Background()

	// Get authorized chats
	chats, err := b.DBClient.GetAuthorizedChats(ctx)
	if err != nil {
		slog.Error("error getting chats", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при получении списка чатов")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	if len(chats) == 0 {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Список авторизованных чатов пуст")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	text := "Авторизованные чаты:\n"
	for _, chat := range chats {
		text += fmt.Sprintf("• %s\n", chat)
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func generateInvite(b *Bot, msg *tgbotapi.Message) {
	if !b.isUserAdmin(msg.From.ID) {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "У вас нет прав для выполнения этой команды")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		slog.Error("error generating random token", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при генерации токена")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}
	token := hex.EncodeToString(bytes)

	ctx := context.Background()
	err := b.DBClient.CreateInviteToken(ctx, token)
	if err != nil {
		slog.Error("error creating invite token", "err", err)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при создании токена")
		msgConfig.ReplyToMessageID = msg.MessageID
		b.sendMessage(msgConfig)
		return
	}

	botUsername := b.TGBotAPI.Self.UserName
	inviteLink := fmt.Sprintf("https://t.me/%s?start=%s", botUsername, token)

	text := fmt.Sprintf("Ссылка для авторизации (действительна 24 часа):\n%s", inviteLink)
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyToMessageID = msg.MessageID
	b.sendMessage(msgConfig)
}

func handleStartCommand(b *Bot, msg *tgbotapi.Message) {
	if !msg.Chat.IsPrivate() {
		return
	}

	token := strings.TrimSpace(msg.CommandArguments())
	if len(token) == 0 {
		unauthorizedResponse := fmt.Sprintf(
			"Добро пожаловать! Для получения доступа обратитесь к администратору. "+
				"Chat ID: %d",
			msg.Chat.ID,
		)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, unauthorizedResponse)
		b.sendMessage(msgConfig)
		return
	}

	ctx := context.Background()
	valid, err := b.DBClient.ValidateInviteToken(ctx, token)
	if err != nil {
		slog.Error("error validating invite token", "err", err, "token", token)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при проверке токена")
		b.sendMessage(msgConfig)
		return
	}

	if !valid {
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Неверный или истекший токен авторизации")
		b.sendMessage(msgConfig)
		return
	}

	err = b.DBClient.AddChat(ctx, msg.Chat.ID)
	if err != nil {
		slog.Error("error adding chat via invite", "err", err, "chat_id", msg.Chat.ID)
		msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Ошибка при авторизации")
		b.sendMessage(msgConfig)
		return
	}

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, "Авторизация успешна! Теперь у вас есть доступ к боту.")
	b.sendMessage(msgConfig)
}
