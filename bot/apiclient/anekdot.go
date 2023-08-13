package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AnecdoteAPI struct {
	HttpClient *http.Client
}

type Anecdote struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}

func (a *AnecdoteAPI) CallAnecdoteApi() (string, error) {
	const max_retry int = 3
	base_url := "https://jokesrv.rubedo.cloud/"
	body := []byte{}

	for i := 1; i <= max_retry; i += 1 {
		resp, err := a.HttpClient.Get(base_url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if resp.StatusCode/100 == 2 {
			break
		} else if i < max_retry {
			slog.Info("got error response from api, retrying...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return "", fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	var an Anecdote
	err := json.Unmarshal(body, &an)
	if err != nil {
		return "", err
	}
	return an.Content, nil
}
