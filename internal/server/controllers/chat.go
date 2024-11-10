package controllers

import (
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type chatControllers interface {
	Create(c echo.Context) error
	GetChats(c echo.Context) error
}

type chat struct {
	store db.Store
}

func (ch chat) Create(c echo.Context) error {
	userID := utils.ExtractUserID(c)

	var chat db.ChatModel
	err := c.Bind(&chat)
	if err != nil || !chat.Valid() {
		return c.JSON(400, nil)
	}
	chat.UserID = userID

	chatID, err := ch.store.Chat.Create(chat)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(201, chatID)
}

func (ch chat) GetChats(c echo.Context) error {
	userID := utils.ExtractUserID(c)
	chats, err := ch.store.Chat.GetChats(userID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, chats)
}

func NewChatControllers(store db.Store) chatControllers {
	return chat{
		store,
	}
}
