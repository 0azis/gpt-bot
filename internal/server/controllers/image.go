package controllers

import (
	"gpt-bot/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type imageControllers interface {
	UploadImage(c echo.Context) error
}
type image struct {
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

func NewImageControllers(savePath string) imageControllers {
	return image{savePath}
}
