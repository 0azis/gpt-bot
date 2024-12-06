package api

import "github.com/rvinnie/yookassa-sdk-go/yookassa"

type yookassaInterface interface {
}

type yookassaClient struct {
	*yookassa.Client
}

func newYookassaClient(token string) yookassaInterface {
	client := yookassa.NewClient("", token)
	return yookassaClient{client}
}
