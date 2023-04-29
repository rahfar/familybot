package apiclient

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
)

const maxPromptSymbolSize = 1000

func CallOpenai(apikey, question string) (string, error) {
	if len(question) > maxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}
	client := openai.NewClient(apikey)
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
