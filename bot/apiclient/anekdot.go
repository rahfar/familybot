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

func (a *AnecdoteAPI) CallAnecdoteAPI() (string, error) {
	const maxRetry = 3
	baseURL := "https://jokesrv.rubedo.cloud/"

	for i := 1; i <= maxRetry; i++ {
		resp, err := a.HttpClient.Get(baseURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if resp.StatusCode/100 == 2 {
			var an Anecdote
			if err := json.Unmarshal(body, &an); err != nil {
				return "", err
			}
			return an.Content, nil
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 3 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(3 * time.Second)
		} else {
			return "", fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return "", fmt.Errorf("max retries reached")
}
