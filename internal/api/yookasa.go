package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type yookassaInterface interface {
	GetAuthLink() string
	GetToken(code string) string
	CreatePayment(access_token string, amount int) bool
}

type yookassaClient struct {
	clientID    string
	secret      string
	redirectURI string
}

var client = &http.Client{}

func newYookassaClient(token string) yookassaInterface {
	return yookassaClient{token, "", "https://api.sponger-code.ru/api/v1/payment/y/webhook"}
}

func (yc yookassaClient) GetAuthLink() string {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://yoomoney.ru/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=account-info", yc.clientID, yc.redirectURI), nil)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	defer resp.Body.Close()

	return req.URL.String()
}

func (yc yookassaClient) GetToken(code string) string {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://yoomoney.ru/oauth/token?code=%s&client_id=%s&grant_type=authorization_code&redirect_uri=%s&client_secret=%s", code, yc.clientID, yc.redirectURI, yc.secret), nil)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		slog.Error(err.Error())
		return ""
	}

	token := struct {
		Value string `json:"access_token"`
	}{}

	err = json.Unmarshal(body, &token)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}

	return token.Value
}

func (yc yookassaClient) CreatePayment(access_token string, amount int) bool {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://yoomoney.ru/api/request-payment?pattern_id=p2p&to=%s&amount_due=%d", "", amount), nil)
	if err != nil {
		slog.Error(err.Error())
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", access_token))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return false
	}
	defer resp.Body.Close()
	fmt.Println(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error())
		return false
	}

	fmt.Println(string(body))

	req2, err := http.NewRequest("POST", fmt.Sprintf("https://yoomoney.ru/api/process-payment?request_id=%s", ""), nil)
	if err != nil {
		slog.Error(err.Error())
		return false
	}
	req2.Header.Add("Authorization", fmt.Sprintf("Bearer %s", access_token))
	req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return true
}
