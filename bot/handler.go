package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func get_users_current_time(data_dir string) string {
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

func remember_tz(message tgbotapi.Message, data_dir string) {
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
