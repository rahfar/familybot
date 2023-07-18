package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/ugjka/go-tz.v2/tz"

	"github.com/rahfar/familybot/bot/apiclient"
)

const user_timezone_file = "user_timezone.json"

type UserTimezone struct {
	UserID    int64   `json:"user_id"`
	UserName  string  `json:"user_name"`
	TimeZone  string  `json:"timezone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func ping(message tgbotapi.Message) string {
	return "понг"
}

func getUsersCurrentTime(data_dir string) string {
	resp := ""
	file_path := data_dir + "/" + user_timezone_file

	file, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}
	user_timezones := make([]UserTimezone, 0, 10)
	err = json.Unmarshal(content, &user_timezones)
	if err != nil {
		log.Printf("[ERROR] Could not unmarshal json: %+v", err)
	}
	for _, ut := range user_timezones {
		loc, err := time.LoadLocation(ut.TimeZone)
		if err != nil {
			log.Printf("[ERROR] Wrong timezone for user: %+v", ut)
			continue
		}
		resp += fmt.Sprintf("У %s сейчас %s\n", ut.UserName, time.Now().In(loc).Format("15:04 MST"))
	}
	if resp == "" {
		resp = "Никто не поделился своим местоположением :("
	}
	return resp
}

func rememberTZ(message tgbotapi.Message, data_dir string) {
	file_path := data_dir + "/" + user_timezone_file
	tz, err := tz.GetZone(tz.Point{Lon: message.Location.Longitude, Lat: message.Location.Latitude})
	if err != nil {
		return
	}

	file, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}
	user_timezones := make([]UserTimezone, 0, 10)
	json.Unmarshal(content, &user_timezones)
	found_flg := false
	for i, ut := range user_timezones {
		if message.From.ID == ut.UserID {
			found_flg = true
			user_timezones[i].TimeZone = tz[0]
			user_timezones[i].Latitude = message.Location.Latitude
			user_timezones[i].Longitude = message.Location.Longitude
			user_timezones[i].UserName = message.From.FirstName
		}
	}
	if !found_flg {
		user_timezones = append(user_timezones, UserTimezone{
			UserID:    message.From.ID,
			TimeZone:  tz[0],
			UserName:  message.From.FirstName,
			Latitude:  message.Location.Latitude,
			Longitude: message.Location.Longitude,
		})
	}
	content, err = json.Marshal(user_timezones)
	if err != nil {
		log.Panic()
	}
	_, err = file.Write(content)
	if err != nil {
		log.Printf("[ERROR] Could not write to file: %+v", err)
	}
}

func getCurrentWeather(w *apiclient.WeatherAPI) string {
	resp := ""
	weather := w.GetWeather()
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].Current.Temp < weather[j].Current.Temp
	})
	if len(weather) > 0 {
		for _, w := range weather {
			resp += fmt.Sprintf("%s: %+g°C (max: %+g°C, min: %+g°C), %s\n", w.Location.Name, w.Current.Temp, w.Forecast.Forecastday[0].Day.Maxtemp_c, w.Forecast.Forecastday[0].Day.Mintemp_c, w.Current.Condition.Text)
		}
		return resp
	} else {
		return "Нет данных"
	}
}

func askChatGPT(o *apiclient.OpenaiAPI, question string) string {
	question = strings.TrimSpace(question)
	resp, err := o.CallGPT3dot5(question)
	if err != nil || len(resp) == 0 {
		log.Printf("[ERROR] Error occured while call openai: %v", err)
		return "Ошибка при вызове ChatGPT :("
	}
	return resp
}

func getYesterdaySales(s *apiclient.SheetsAPI) string {
	yesterday := time.Now().Add(-24 * time.Hour)
	_, month_total, err := s.CallGoogleSheetsApi(yesterday.Day(), int(yesterday.Month()))
	if err != nil {
		return "Возникла ошибка при чтении данных :("
	}
	return fmt.Sprintf("Продажи с начала мес: %.2f₽\n", month_total)
}

func getAnecdote(a *apiclient.AnecdoteAPI) string {
	anecdote, err := a.CallAnecdoteApi()
	if err != nil {
		log.Printf("[ERROR] error calling anecdote api: %v", err)
		return "Не смог получить свежий анекдот :("
	}
	return anecdote
}

func getLatestNews(k *apiclient.KommersantAPI) string {
	news, err := k.CallKommersantAPI()
	if (err != nil) || (len(news) == 0) {
		log.Printf("[ERROR] error calling news api: %v", err)
		return "Не смог получить последние новости :("
	}
	resp := "\nПоследние новости:\n"
	for i, n := range news[:3] {
		resp += fmt.Sprintf("%d. [%s](%s)\n", i+1, n.Title, n.Link)
	}
	return resp
}

func transcriptVoice(s *apiclient.OpenaiAPI, b *tgbotapi.BotAPI, FileID string) string {
	
	// Get direct link to audio message
	link, err := b.GetFileDirectURL(FileID)
	if err != nil {
		log.Printf("[ERROR] getting voice msg: %v", err)
		return "Ошибка при скачивании голосового сообщения"
		
	}
	filename := "/tmp/" + FileID + path.Ext(link)
	log.Printf("[DEBUG] audio filename: %s", filename)
	// Download audio file
	resp, err := http.Get(link)
	if err != nil {
		log.Printf("[ERROR] getting voice msg: %v", err)
		return "Ошибка при скачивании голосового сообщения"
	}
	defer resp.Body.Close()

	// Create the output file
	file, err := os.Create(filename)
	if err != nil {
		return "Ошибка при скачивании голосового сообщения"
	}
	defer file.Close()
	defer os.Remove(filename)

	// Write the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "Ошибка при скачивании голосового сообщения"
	}

	text, err := s.CallWhisper(filename)
	if err != nil {
		log.Printf("[ERROR] getting voice msg: %v", err)
		return "Ошибка при обработки голосового сообщения"
	}
	return text
}
