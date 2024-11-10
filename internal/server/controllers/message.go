package controllers

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

type messageControllers interface {
	NewMessage(c echo.Context) error
	GetMessages(c echo.Context) error
}

type message struct {
	api   api.ApiInterface
	store db.Store
}

func (m message) NewMessage(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	var msgCredentials db.MessageCredentials
	err := c.Bind(&msgCredentials)
	if err != nil {
		return c.JSON(400, nil)
	}

	userMsg := db.NewUserMessage(msgCredentials.ChatID, msgCredentials.Content)

	err = m.store.Message.Create(userMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	model, err := m.store.Chat.GetModelOfChat(msgCredentials.ChatID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	messages, err := m.store.Message.GetMessages(jwtUserID, msgCredentials.ChatID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	answer, err := m.api.SendMessage(model, messages)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	assistantMsg := db.NewAssistantMessage(msgCredentials.ChatID, answer)
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

func NewMessageControllers(api api.ApiInterface, store db.Store) messageControllers {
	return message{
		api, store,
	}
}
