package controllers

import (
	"gpt-bot/internal/db"
	"gpt-bot/utils"

	"github.com/labstack/echo/v4"
)

type userControllers interface {
	GetUser(c echo.Context) error
}

type user struct {
	store db.Store
}

// GetUser returns a user by access token
func (u user) GetUser(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	user, err := u.store.User.GetUser(jwtUserID)
	if err != nil {
		return c.JSON(500, nil)
	}

	return c.JSON(200, user)
}

// import all controllers
func NewUserControllers(store db.Store) userControllers {
	return user{store}
}
