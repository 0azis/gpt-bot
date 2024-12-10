package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type yookassaInterface interface {
	Auth()
}

type b struct {
	clientID     string
	responseType string
	redirectURI  string
	scope        string
	instanceName string
}

type yookassaClient struct {
	token string
}

func newYookassaClient(token string) yookassaInterface {
	return yookassaClient{token}
}

func (yc yookassaClient) Auth() {
	data := url.Values{}
	data.Set("client_id", yc.token)
	data.Set("redirect_uri", "https://shit.ru/redirect")
	data.Set("response_type", "code")
	data.Set("scope", "account-info")

	req, err := http.NewRequest("POST", fmt.Sprintf("https://yoomoney.ru/oauth/authorize?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s", data.Get("client_id"), data.Get("response_type"), data.Get("redirect_uri"), data.Get("scope")), nil)
	if err != nil {
		slog.Error(err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{
		CheckRedirect: checkRedirect,
	}

	resp, err := c.Do(req)
	if err != nil {
		slog.Error(err.Error())
	}

	// fmt.Println(resp.
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Ошибка при закрытии тела ответа:", err)
		}
	}()

	// c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
	// 	return errors.New("Redirect")
	// }

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body), err)
	// fmt.Println("BODY", string(res))
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	fmt.Println(via[0])
	fmt.Println("REDIRECT", req)

	// c := &http.Client{
	// 	CheckRedirect: checkRedirect,
	// }
	// resp, err := c.Do(req)
	// if err != nil {
	// 	slog.Error(err.Error())
	// }

	return nil
}
