package server

import (
	"gpt-bot/utils"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		bearer := c.Request().Header.Get("Authorization")
		token := strings.Split(bearer, " ")
		if len(token) == 1 {
			return c.JSON(401, nil)
		}
		_, err := utils.GetIdentity(token[1])
		if err != nil {
			return c.JSON(401, nil)
		}
		return next(c)
	}
}
