package api

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type ApiInterface interface {
	SendMessage(message string) (string, error)
}

type apiClient struct {
	*openai.Client
}

func New(apiToken string) ApiInterface {
	client := openai.NewClient(apiToken)
	return apiClient{
		client,
	}
}

func (ac apiClient) SendMessage(message string) (string, error) {
	resp, err := ac.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleAssistant, Content: message},
		},
	})

	fmt.Println(resp)
	return resp.Choices[0].Message.Content, err
}
