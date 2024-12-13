package controllers

import (
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

type chatControllers interface {
	GetChats(c echo.Context) error
	Delete(c echo.Context) error
}

type chat struct {
	store db.Store
}

func (ch chat) GetChats(c echo.Context) error {
	userID := utils.ExtractUserID(c)
	chats, err := ch.store.Chat.GetByUser(userID)
	for _, chat := range chats {
		chat.SetModelID()
	}

	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, chats)
}

func (ch chat) Delete(c echo.Context) error {
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(400, nil)
	}

	err = ch.store.Chat.Delete(chatID)
	if err != nil {
		return c.JSON(500, nil)
	}
	return c.JSON(200, nil)
}

func NewChatControllers(store db.Store) chatControllers {
	return chat{
		store,
	}
}
