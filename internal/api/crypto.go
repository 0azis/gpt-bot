package api

// const baseUrl = ""

type cryptoInterface interface {
	CreateInvoiceLink() (string, error)
}

type cryptoClient struct {
	token string
}

type invoicePayload struct {
	Amount    int    `json:"amount"`
	Asset     string `json:"asset"`
	Payload   string `json:"payload"`
	ExpiresIn int    `json:"-"`
}

func newCrypto(token string) cryptoInterface {
	return cryptoClient{token}
}

func (c cryptoClient) CreateInvoiceLink() (string, error) {
	return "", nil
}
