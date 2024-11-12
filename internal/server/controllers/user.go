package controllers

import (
	"database/sql"
	"errors"
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type userControllers interface {
	GetUser(c echo.Context) error
}

type user struct {
	store db.Store
}

func (u user) GetUser(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	user, err := u.store.User.GetByID(jwtUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(404, nil)
	}
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, user)
}

func NewUserControllers(store db.Store) userControllers {
	return user{store}
}
