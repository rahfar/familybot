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
	Role      string
	Content   string
	ImageData string // base64 encoded image
	Time      time.Time
}

type OpenaiAPI struct {
	ApiKey     string
	HttpClient *http.Client
}

const MaxPromptSymbolSize = 4096

func (o *OpenaiAPI) requestChatCompletion(messages []openai.ChatCompletionMessage, model string) (string, error) {
	const maxRetry = 3
	const defaultModel = "gpt-5-mini"
	if model == "" {
		model = defaultModel
	}
	for i := 1; i <= maxRetry; i++ {
		client := openai.NewClient(o.ApiKey)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
			},
		)
		if err == nil {
			return resp.Choices[0].Message.Content, nil
		}
		if APIError, ok := err.(*openai.APIError); ok && i < maxRetry {
			slog.Info(
				"got error response from api, retrying in 5 seconds...",
				"retry-cnt", i,
				"status", APIError.HTTPStatusCode,
			)
			time.Sleep(5 * time.Second)
		} else {
			return "", err
		}
	}
	return "", fmt.Errorf("max retries reached")
}

func (o *OpenaiAPI) GenerateChatCompletion(question string, imageData string, responseHistory []GPTResponse) (string, error) {
	if len(question) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	for _, v := range responseHistory {
		if v.ImageData != "" {
			// Message with image - create data URL from base64
			dataURL := "data:image/jpeg;base64," + v.ImageData
			messages = append(messages, openai.ChatCompletionMessage{
				Role: v.Role,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: v.Content,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    dataURL,
							Detail: openai.ImageURLDetailAuto,
						},
					},
				},
			})
		} else {
			// Text-only message
			messages = append(messages, openai.ChatCompletionMessage{Role: v.Role, Content: v.Content})
		}
	}

	// Add current message
	if imageData != "" {
		// Message with image - create data URL from base64
		dataURL := "data:image/jpeg;base64," + imageData
		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: question,
				},
				{
					Type:     openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{URL: dataURL, Detail: openai.ImageURLDetailAuto},
				},
			},
		})
	} else {
		// Text-only message
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: question})
	}

	return o.requestChatCompletion(messages, "gpt-5")
}

func (o *OpenaiAPI) CorrectGrammarAndStyle(text string) (string, error) {
	gptcontext := "Correct the following English text for grammar, punctuation, " +
		"spelling and capitalization while preserving the original meaning and tone. " +
		"Return only the corrected sentence(s)."

	if len(text) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: gptcontext},
		{Role: openai.ChatMessageRoleUser, Content: "Input: " + text},
	}

	return o.requestChatCompletion(messages, "gpt-5-nano")
}

func (o *OpenaiAPI) TranslateEnglishToRussian(text string) (string, error) {
	prompt := "translate from english to russian: " + text

	if len(prompt) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}

	return o.requestChatCompletion(messages, "gpt-5-nano")
}

func (o *OpenaiAPI) TranslateRussianToEnglish(text string) (string, error) {
	prompt := "переведи с русского на английский: " + text

	if len(prompt) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}

	return o.requestChatCompletion(messages, "gpt-5-nano")
}

func (o *OpenaiAPI) TranscribeAudioFile(filePath string) (string, error) {
	const maxRetry = 3

	c := openai.NewClient(o.ApiKey)
	ctx := context.Background()

	for i := 1; i <= maxRetry; i++ {
		req := openai.AudioRequest{
			Model:    "gpt-4o-transcribe",
			FilePath: filePath,
			Format:   openai.AudioResponseFormatJSON,
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

func (o *OpenaiAPI) GenerateImageWithPrompt(prompt string) (string, error) {
	const maxRetry = 3
	c := openai.NewClient(o.ApiKey)
	ctx := context.Background()
	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
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
