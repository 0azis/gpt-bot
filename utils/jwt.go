package utils

import (
	"gpt-bot/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	TOKEN_TIME_ACCESS  int64 = 1000
	TOKEN_TIME_REFRESH int64 = 432000
)

type tokenInterface interface {
	// getter
	GetUserID() int
	GetStrToken() string

	// setter
	SetUserID(userID int)
	SetStrToken(token string)

	// methods
	SignJWT() error
	GetIdentity() error
}

type token struct {
	userID         int
	str            string
	secretHash     []byte
	expirationTime int64
}

func NewToken() tokenInterface {
	return &token{
		secretHash: config.JwtSecretHash(),
	}
}

func (t *token) GetUserID() int {
	return t.userID
}

func (t *token) GetStrToken() string {
	return t.str
}

func (t *token) SetUserID(userID int) {
	t.userID = userID
}

func (t *token) SetStrToken(token string) {
	t.str = token
}

func (t *token) SignJWT() error {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Создаем payload структуру
		"id": t.userID,                                     // UserId для идентификации пользователя
		"e":  int64(time.Now().Unix()) + TOKEN_TIME_ACCESS, // expiredTime для безопасности
	}).SignedString(t.secretHash)
	t.str = token

	return err
}

func (t *token) GetIdentity() error {
	identity, err := jwt.Parse(t.str, func(token *jwt.Token) (interface{}, error) {
		return t.secretHash, nil
	})

	if err != nil {
		return err
	}

	values := identity.Claims.(jwt.MapClaims)
	userId := int(values["id"].(float64))
	expiredTime := int64(values["e"].(float64))

	t.userID = userId
	t.expirationTime = expiredTime

	return nil
}
