package apiclient

import (
	"context"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
)

type OpenaiAPI struct {
	ApiKey     string
	HttpClient *http.Client
}

const maxPromptSymbolSize = 1000

func (o *OpenaiAPI) CallGPT3dot5(question string) (string, error) {
	if len(question) > maxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}
	client := openai.NewClient(o.ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenaiAPI) CallWhisper(filePath string) (string, error) {
	c := openai.NewClient(o.ApiKey)
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filePath,
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
