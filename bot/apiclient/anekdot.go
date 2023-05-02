package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Anecdote struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}

func CallAnecdoteApi() (string, error) {
	base_url := "https://jokesrv.rubedo.cloud/"
	resp, err := http.Get(base_url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("non 2** HTTP status code: %d - %s - %s", resp.StatusCode, resp.Status, string(body))
	}
	var a Anecdote
	err = json.Unmarshal(body, &a)
	if err != nil {
		return "", err
	}
	return a.Content, nil
}
