package utils

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"
)

func ExtractUserID(c echo.Context) int {
	authorizationHeader := c.Request().Header.Get("Authorization")
	token := strings.Split(authorizationHeader, " ")[1]
	payload, _ := GetIdentity(token)
	return payload.UserID
}

func BindJSON(c echo.Context, i interface{}) error {
	err := json.NewDecoder(c.Request().Body).Decode(i)
	return err
}
