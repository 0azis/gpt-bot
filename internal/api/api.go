package api

import (
	"context"
	"gpt-bot/internal/db"

	"github.com/sashabaranov/go-openai"
)

type ApiInterface interface {
	SendMessage(apiMsg db.MessageModel) (string, error)
}

type apiClient struct {
	*openai.Client
}

type apiMessageCredentials struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func New(apiToken string) ApiInterface {
	client := openai.NewClient(apiToken)
	return apiClient{
		client,
	}
}

func (ac apiClient) SendMessage(apiMsg db.MessageModel) (string, error) {
	resp, err := ac.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: apiMsg.Model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleAssistant, Content: apiMsg.Content},
		},
	})

	if len(resp.Choices) == 0 {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
