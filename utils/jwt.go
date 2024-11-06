package utils

import (
	"gpt-bot/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	SECRET_HASH        []byte = config.JwtSecretHash()
	TOKEN_TIME_ACCESS  int64  = 1000
	TOKEN_TIME_REFRESH int64  = 432000
)

type tokenPayload struct {
	UserID         int
	expirationTime int64
}

func SignJWT(userId int) (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Создаем payload структуру
		"id": userId,                                       // UserId для идентификации пользователя
		"e":  int64(time.Now().Unix()) + TOKEN_TIME_ACCESS, // expiredTime для безопасности
	}).SignedString(SECRET_HASH)
	return token, err
}

func GetIdentity(token string) (tokenPayload, error) {
	var jwtPayload tokenPayload
	identity, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return SECRET_HASH, nil
	})

	if err != nil {
		return jwtPayload, err
	}

	values := identity.Claims.(jwt.MapClaims)
	userId := int(values["id"].(float64))
	expiredTime := int64(values["e"].(float64))

	jwtPayload.UserID = userId
	jwtPayload.expirationTime = expiredTime

	return jwtPayload, nil
}
