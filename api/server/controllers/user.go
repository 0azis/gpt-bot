package controllers

import (
	"gpt-bot/api/db"

	"github.com/labstack/echo/v4"
)

type userControllers interface {
	Hello(c echo.Context) error
}

type user struct {
	store db.Store
}

// Hello prints Hello string (test)
func (u user) Hello(c echo.Context) error {
	c.JSON(200, "Hello")
	return nil
}

// import all controllers
func NewUserControllers(store db.Store) userControllers {
	return user{store}
}
