package apiclient

import (
	"fmt"
	"testing"
)

func TestGetWeather(t *testing.T) {
	apikey := ""
	city := []string{""}
	result := GetWeather(apikey, city)
	fmt.Printf("%+v", result)
}
