package utils

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func ExtractUserID(c echo.Context) int {
	authorizationHeader := c.Request().Header.Get("Authorization")
	token := strings.Split(authorizationHeader, " ")[1]

	t := NewToken()
	t.SetStrToken(token)
	err := t.GetIdentity()
	if err != nil {
		return 0
	}

	return t.GetUserID()
}
