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

// type JWT struct {
// 	Access  string
// 	Refresh string
// }

func SignJWT(userId int) (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Создаем payload структуру
		"id": userId,                                       // UserId для идентификации пользователя
		"e":  int64(time.Now().Unix()) + TOKEN_TIME_ACCESS, // expiredTime для безопасности
	}).SignedString(SECRET_HASH)
	return token, err
}

// func createRefreshToken(userId int) (string, error) {
// 	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		// Создаем payload структуру
// 		"id": userId,                                        // UserId для идентификации пользователя
// 		"e":  int64(time.Now().Unix()) + TOKEN_TIME_REFRESH, // expiredTime для безопасности
// 	}).SignedString(SECRET_HASH)
// 	return token, err
// }

// func NewJWT(userID int) (JWT, error) {
// 	var jwtResult JWT
// 	accessToken, err := createAccessToken(userID)
// 	if err != nil {
// 		return jwtResult, err
// 	}
// 	refrestToken, err := createRefreshToken(userID)
// 	if err != nil {
// 		return jwtResult, err
// 	}

// 	jwtResult.Access = accessToken
// 	jwtResult.Refresh = refrestToken
// 	return jwtResult, nil
// }

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

// func IsValid(payload tokenPayload) bool {
// 	return payload.expirationTime > time.Now().Unix()
// }
