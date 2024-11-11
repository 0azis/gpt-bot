package controllers

import (
	"errors"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

type messageControllers interface {
	NewMessage(c echo.Context) error
	ChatMessage(c echo.Context) error
	GetMessages(c echo.Context) error

	modelAnswer(chat db.ChatModel, msg []db.MessageModel) (string, error)
}

type message struct {
	api   api.Interface
	store db.Store
}

func (m message) NewMessage(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	model := c.QueryParam("model")
	if model == "" {
		return c.JSON(400, nil)
	}

	var msgCredentials db.MessageCredentials
	err := c.Bind(&msgCredentials)
	if err != nil || !msgCredentials.Valid() {
		return c.JSON(400, nil)
	}

	newChat := db.ChatModel{
		UserID: jwtUserID,
		Model:  model,
	}
	if ok := newChat.SetType(); !ok {
		return c.JSON(400, nil)
	}

	chatID, err := m.store.Chat.Create(newChat)
	if err != nil {
		return c.JSON(500, nil)
	}
	newChat.ID = chatID

	userMsg := db.NewUserMessage(chatID, msgCredentials.Content)

	go func() {
		title, _ := m.api.OpenAI.GenerateTopicForChat(userMsg)
		m.store.Chat.UpdateTitle(chatID, title)
	}()

	err = m.store.User.ReduceBalance(jwtUserID, db.PriceOfMessage)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(500, nil)
	}

	answer, err := m.modelAnswer(newChat, []db.MessageModel{userMsg})
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = m.store.Message.Create(userMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	assistantMsg := db.NewAssistantMessage(chatID, answer)

	err = m.store.Message.Create(assistantMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, chatID)
}

func (m message) ChatMessage(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	value := c.Param("id")
	chatID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	var msgCredentials db.MessageCredentials
	err = c.Bind(&msgCredentials)
	if err != nil || !msgCredentials.Valid() {
		return c.JSON(400, nil)
	}

	chat, err := m.store.Chat.GetChatInfo(chatID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	userMsg := db.NewUserMessage(chat.ID, msgCredentials.Content)

	messages, err := m.store.Message.GetMessages(jwtUserID, chat.ID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = m.store.User.ReduceBalance(jwtUserID, db.PriceOfMessage)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	answer, err := m.modelAnswer(chat, messages)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = m.store.Message.Create(userMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	assistantMsg := db.NewAssistantMessage(chat.ID, answer)
	err = m.store.Message.Create(assistantMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, answer)
}

func (m message) GetMessages(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	value := c.Param("id")
	chatID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	messages, err := m.store.Message.GetMessages(jwtUserID, chatID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, messages)
}

func (m message) modelAnswer(chat db.ChatModel, msg []db.MessageModel) (string, error) {
	switch chat.Type {
	case db.ChatImage:
		switch chat.Model {
		case "runware":
			answer, err := m.api.Runware.SendMessage(msg[0].Content)
			if err != nil {
				slog.Error(err.Error())
				return answer, err
			}
			return answer, nil
		case "dall-e-3":
			answer, err := m.api.OpenAI.SendImageMessage(msg[0].Content)
			if err != nil {
				slog.Error(err.Error())
				return answer, err
			}
			return answer, nil
		default:
			return "", errors.New("model error")
		}

	case db.ChatText:
		answer, err := m.api.OpenAI.SendMessage(chat.Model, msg)
		if err != nil {
			slog.Error(err.Error())
			return answer, err
		}
		return answer, nil
	default:
		return "", errors.New("model error")
	}
}

func NewMessageControllers(api api.Interface, store db.Store) messageControllers {
	return message{
		api, store,
	}
}
