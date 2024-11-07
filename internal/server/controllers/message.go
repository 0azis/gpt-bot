package controllers

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
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
	var userMessage db.MessageModel
	err := c.Bind(&userMessage)
	if err != nil {
		return c.JSON(400, nil)
	}
	userMessage.IsUser = true

	err = m.store.Message.Create(userMessage)
	if err != nil {
		return c.JSON(500, nil)
	}

	model, err := m.store.Chat.GetModelOfChat(userMessage.ChatID)
	if err != nil {
		return c.JSON(500, nil)
	}
	userMessage.Model = model

	answer, err := m.api.SendMessage(userMessage)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	botMsg := db.MessageModel{
		ChatID:  userMessage.ChatID,
		Content: answer,
	}
	err = m.store.Message.Create(botMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, answer)
}

func (m message) GetMessages(c echo.Context) error {
	value := c.Param("id")
	chatID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	messages, err := m.store.Message.GetMessages(chatID)
	if err != nil {
		return c.JSON(500, nil)
	}

	return c.JSON(200, messages)
}

func NewMessageControllers(api api.ApiInterface, store db.Store) messageControllers {
	return message{
		api, store,
	}
}
