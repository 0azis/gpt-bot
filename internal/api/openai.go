package api

import (
	"context"
	"gpt-bot/internal/db"
	"log/slog"

	"github.com/sashabaranov/go-openai"
)

type openAiInterface interface {
	SendMessage(model string, apiMsgs []db.MessageModel) (string, error)
	SendImageMessage(apiMsgs []db.MessageModel) (string, error)
}

type apiClient struct {
	*openai.Client
}

type apiMessageCredentials struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func newOpenAiClient(token string) openAiInterface {
	client := openai.NewClient(token)
	return apiClient{
		client,
	}
}

func (ac apiClient) SendMessage(model string, apiMsgs []db.MessageModel) (string, error) {
	var openaiMessages []openai.ChatCompletionMessage
	for _, message := range apiMsgs {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role: message.Role, Content: message.Content,
		})
	}
	resp, err := ac.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
	})

	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (ac apiClient) SendImageMessage(apiMsgs []db.MessageModel) (string, error) {
	var openaiMessages []openai.ChatCompletionMessage
	for _, message := range apiMsgs {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role: message.Role, Content: message.Content,
		})
	}

	resp, err := ac.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.CreateImageModelDallE3,
		Messages: openaiMessages,
	})

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
