package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rahfar/familybot/bot/apiclient"
	"gopkg.in/ugjka/go-tz.v2/tz"
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

func getCurrentWeather(apikey string, cities []string) string {
	resp := ""
	weather := apiclient.GetWeather(apikey, cities)
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

func askChatGPT(apikey, question string) string {
	question = strings.TrimSpace(question)
	resp, err := apiclient.CallOpenai(apikey, question)
	if err != nil || len(resp) == 0 {
		log.Printf("[ERROR] Error occured while call openai: %v", err)
		return "Ошибка при вызове ChatGPT :("
	}
	return resp
}

func getYesterdaySales(apikey, spreadsheetid string) string {
	yesterday := time.Now().Add(-24 * time.Hour)
	sales, month_total, err := apiclient.CallGoogleSheetsApi(apikey, spreadsheetid, yesterday.Day(), int(yesterday.Month()))
	total := 0.0
	if err != nil {
		return "Возникла ошибка при чтении данных :("
	}
	resp := "Продажи за вчера:\n"
	for _, sale := range sales {
		total += sale.SalesValue
		resp += fmt.Sprintf("    %s - %s - %.2f₽\n", sale.Name, sale.SalesType, sale.SalesValue)
	}
	resp += fmt.Sprintf("Итого: %.2f₽\n", total)
	resp += fmt.Sprintf("Итого с начала мес: %.2f₽\n", month_total)
	return resp
}

func getAnecdote() string {
	anecdote, err := apiclient.CallAnecdoteApi()
	if err != nil {
		log.Printf("[ERROR] error calling anecdote api: %v", err)
		return "Не смог получить свежий анекдот :("
	}
	return anecdote
}

func getLatestNews() string {
	news, err := apiclient.CallKommersantAPI()
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
