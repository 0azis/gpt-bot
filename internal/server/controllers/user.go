package controllers

import (
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
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	if user.ID == 0 {
		return c.JSON(404, nil)
	}

	go func() {
		u.store.User.SetIsNewFalse(user.ID)
		u.store.Stat.Count(int64(jwtUserID))
	}()

	return c.JSON(200, user)
}

func NewUserControllers(store db.Store) userControllers {
	return user{store}
}
