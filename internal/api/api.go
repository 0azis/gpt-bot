package api

import (
	"gpt-bot/config"
)

type Interface struct {
	OpenAI   openAiInterface
	Runware  runwareInterface
	Crypto   cryptoInterface
	YooKassa yookassaInterface
}

func New(cfg config.Api) Interface {
	openai := newOpenAiClient(cfg.OpenAI())
	runware := newRunware(cfg.Runware())
	crypto := newCrypto(cfg.CryptoBot())
	yookassa := newYookassaClient(cfg.YooKassa())

	return Interface{
		openai, runware, crypto, yookassa,
	}
}
