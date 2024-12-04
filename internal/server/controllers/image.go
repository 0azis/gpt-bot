package controllers

import (
	"gpt-bot/tgbot"
	"gpt-bot/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type imageControllers interface {
	UploadImage(c echo.Context) error
	SendImageToTelegram(c echo.Context) error
}
type image struct {
	bot      tgbot.BotInterface
	savePath string
}

func (i image) UploadImage(c echo.Context) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(400, nil)
	}

	filename, err := utils.SaveImage(file, i.savePath)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, filename)
}

func (i image) SendImageToTelegram(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	imageLink := struct {
		Link string `json:"imageLink"`
	}{}
	if err := c.Bind(&imageLink); err != nil {
		return c.JSON(400, nil)
	}
	if imageLink.Link == "" {
		return c.JSON(400, nil)
	}

	i.bot.SendImage(imageLink.Link, jwtUserID)
	return c.JSON(200, nil)
}

func NewImageControllers(savePath string, bot tgbot.BotInterface) imageControllers {
	return image{bot, savePath}
}
