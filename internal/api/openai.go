package api

import (
	"context"
	"gpt-bot/internal/db"
	"log/slog"

	"github.com/sashabaranov/go-openai"
)

type openAiInterface interface {
	SendMessage(model string, apiMsgs []db.MessageModel) (string, error)
	SendImageMessage(prompt string) (string, error)
	GenerateTopicForChat(startMsg db.MessageModel) (string, error)
}

type openaiClient struct {
	*openai.Client
}

type apiMessageCredentials struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func newOpenAiClient(token string) openAiInterface {
	client := openai.NewClient(token)
	return openaiClient{
		client,
	}
}

func (oc openaiClient) SendMessage(model string, apiMsgs []db.MessageModel) (string, error) {
	var openaiMessages []openai.ChatCompletionMessage
	for _, message := range apiMsgs {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role: message.Role, Content: message.Content,
		})
	}
	resp, err := oc.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
	})

	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (oc openaiClient) SendImageMessage(prompt string) (string, error) {
	resp, err := oc.CreateImage(context.Background(), openai.ImageRequest{
		Model:  openai.CreateImageModelDallE3,
		Prompt: prompt,
	})

	if err != nil {
		return "", err
	}

	return resp.Data[0].URL, nil
}

func (oc openaiClient) GenerateTopicForChat(startMsg db.MessageModel) (string, error) {
	resp, err := oc.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: startMsg.Content},
			{Role: openai.ChatMessageRoleUser, Content: "generate a short chat topic (5 words max) based on the initial message in Russian"},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, err
}
