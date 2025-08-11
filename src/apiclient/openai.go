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

const MaxPromptSymbolSize = 4096

func (o *OpenaiAPI) callChatCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	const maxRetry = 3

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

func (o *OpenaiAPI) CallGPT(question string, responseHistory []GPTResponse) (string, error) {
	if len(question) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	for _, v := range responseHistory {
		messages = append(messages, openai.ChatCompletionMessage{Role: v.Role, Content: v.Content})
	}

	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: question})
	return o.callChatCompletion(messages)
}

func (o *OpenaiAPI) CallGPTforEng(text string) (string, error) {
	gptcontext := "Act as an expert in English language arts with advanced experience in proofreading, editing, spelling, grammar, proper sentence structure, and punctuation. You have critical thinking skills with the ability to analyze and evaluate information, arguments, and ideas, and to make logical and well-supported judgments and decisions. You will be provided content from a professional business to proofread in the form of emails, texts, and instant messages to make sure they are error-free before sending. Your approach would be to carefully read through each communication to identify any errors, inconsistencies, or areas where clarity could be improved. Your overall goal is to ensure communications are error-free, clear, and effective in achieving their intended purpose. You will make appropriate updates to increase readability, professionalism, and cohesiveness, while also ensuring that your intended meaning is conveyed accurately. Only reply to the correction, and the improvements, and nothing else, do not write explanations."

	if len(text) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: gptcontext},
		{Role: openai.ChatMessageRoleUser, Content: "Fix English: " + text},
	}

	return o.callChatCompletion(messages)
}

func (o *OpenaiAPI) CallGPTEng2Ru(text string) (string, error) {
	prompt := "translate from english to russian: " + text

	if len(prompt) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}

	return o.callChatCompletion(messages)
}

func (o *OpenaiAPI) CallGPTRu2Eng(text string) (string, error) {
	prompt := "переведи с русского на английский: " + text

	if len(prompt) > MaxPromptSymbolSize {
		return "Слишком длинный вопрос, попробуйте покороче", nil
	}

	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}

	return o.callChatCompletion(messages)
}

func (o *OpenaiAPI) CallTranscriptionEndpoint(filePath string) (string, error) {
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

func (o *OpenaiAPI) CallDalle(prompt string) (string, error) {
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
