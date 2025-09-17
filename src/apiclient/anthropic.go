package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AnthropicAPI struct {
	ApiKey      string
	HttpClient  *http.Client
	Model       string
	ApiVersion  string
	MaxTokens   int
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Response struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	Content      []Content `json:"content"`
	StopReason   string    `json:"stop_reason"`
	StopSequence *string   `json:"stop_sequence,omitempty"`
	Usage        Usage     `json:"usage"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Payload struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

func (a *AnthropicAPI) CallGPT(question string, responseHistory []GPTResponse) (string, error) {
	const maxRetry = 3
	var response Response
	url := "https://api.anthropic.com/v1/messages"

	if len(question) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := make([]Message, 0)
	for _, v := range responseHistory {
		messages = append(messages, Message{Role: v.Role, Content: v.Content})
	}
	messages = append(messages, Message{Role: "user", Content: question})

	payload := Payload{Model: a.Model, MaxTokens: a.MaxTokens, Messages: messages}
	pl, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error creating request", "err", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(pl))
	if err != nil {
		slog.Error("Error creating request", "err", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.ApiKey)
	req.Header.Set("anthropic-version", a.ApiVersion)

	for i := 1; i <= maxRetry; i++ {
		resp, err := a.HttpClient.Do(req)
		if err != nil {
			slog.Error("Error creating request", "err", err)
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Error creating request", "err", err)
			return "", err
		}

		if resp.StatusCode/100 == 2 {
			err := json.Unmarshal(body, &response)
			if err != nil {
				slog.Error("could not unmarshal json body", "err", err)
				return "", err
			}
			return response.Content[0].Text, nil
		}

		if i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", resp.Status, "body", string(body))
			time.Sleep(5 * time.Second)
		} else {
			return "", fmt.Errorf("got error response from api: %s - %s", resp.Status, string(body))
		}
	}

	return "", fmt.Errorf("max retries reached")
}
