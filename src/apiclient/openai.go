package apiclient

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type GPTResponse struct {
	Role    string
	Content string
	Time    time.Time
}

type OpenaiAPI struct {
	ApiKey     string
	HttpClient *http.Client
	GPTModel   string
}

const maxPromptSymbolSize = 2000

func (o *OpenaiAPI) CallGPT(question string, responseHistory []GPTResponse) (string, error) {
	const maxRetry = 3

	if len(question) > maxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	for _, v := range responseHistory {
		messages = append(messages, openai.ChatCompletionMessage{Role: v.Role, Content: v.Content})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: question})

	for i := 1; i <= maxRetry; i++ {
		client := openai.NewClient(o.ApiKey)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    o.GPTModel,
				Messages: messages,
			},
		)
		if err == nil {
			return resp.Choices[0].Message.Content, nil
		}
		if APIError, ok := err.(*openai.APIError); ok && i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", APIError.HTTPStatusCode)
			time.Sleep(5 * time.Second)
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("max retries reached")
}

func (o *OpenaiAPI) CallWhisper(filePath string) (string, error) {
	const maxRetry = 3

	c := openai.NewClient(o.ApiKey)
	ctx := context.Background()

	for i := 1; i <= maxRetry; i++ {
		req := openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: filePath,
		}
		resp, err := c.CreateTranscription(ctx, req)
		if err == nil {
			return resp.Text, nil
		}
		if APIError, ok := err.(*openai.APIError); ok && i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", APIError.HTTPStatusCode)
			time.Sleep(5 * time.Second)
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("max retries reached")
}

func (o *OpenaiAPI) CallDalle(prompt string) (string, error) {
	const maxRetry = 3
	c := openai.NewClient(o.ApiKey)
	ctx := context.Background()
	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}
	for i := 1; i <= maxRetry; i++ {
		respUrl, err := c.CreateImage(ctx, reqUrl)
		if err == nil {
			return respUrl.Data[0].URL, nil
		}
		if APIError, ok := err.(*openai.APIError); ok && i < maxRetry {
			slog.Info("got error response from api, retrying in 5 seconds...", "retry-cnt", i, "status", APIError.HTTPStatusCode)
			time.Sleep(5 * time.Second)
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("max retries reached")
}
