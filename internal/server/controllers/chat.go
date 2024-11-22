package controllers

import (
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type chatControllers interface {
	GetChats(c echo.Context) error
}

type chat struct {
	store db.Store
}

func (ch chat) GetChats(c echo.Context) error {
	userID := utils.ExtractUserID(c)
	chats, err := ch.store.Chat.GetByUser(userID)
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
