package utils

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func ExtractUserID(c echo.Context) int {
	authorizationHeader := c.Request().Header.Get("Authorization")
	token := strings.Split(authorizationHeader, " ")[1]
	payload, _ := GetIdentity(token)
	return payload.UserID
}
