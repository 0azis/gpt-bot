package api

import (
	"gpt-bot/config"
)

type Interface struct {
	OpenAI  openAiInterface
	Runware runwareInterface
}

func New(cfg config.Api) Interface {
	openai := newOpenAiClient(cfg.OpenAI())
	runware := newRunware(cfg.Runware())
	return Interface{
		openai, runware,
	}
}
