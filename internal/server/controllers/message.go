package controllers

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/utils"

	"github.com/labstack/echo/v4"
)

type messageControllers interface {
	NewMessage(c echo.Context) error
}

type message struct {
	api   api.ApiInterface
	store db.Store
}

type messageCredentials struct {
	Prompt string `json:"prompt"`
}

func (m message) NewMessage(c echo.Context) error {
	userID := utils.ExtractUserID(c)
	var messageCredentials messageCredentials
	err := c.Bind(&messageCredentials)
	answer, err := m.api.SendMessage(messageCredentials.Prompt)
	if err != nil {
		return c.JSON(500, nil)
	}

	err = m.store.Message.Create(userID)
	if err != nil {
		return c.JSON(500, nil)
	}

	err = m.store.Chat.Create(userID)
	if err != nil {
		return c.JSON(500, nil)
	}

	return c.JSON(200, answer)
}

func NewMessageControllers(api api.ApiInterface, store db.Store) messageControllers {
	return message{
		api, store,
	}
}
