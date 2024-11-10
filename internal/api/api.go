package api

import (
	"fmt"
	"gpt-bot/config"
)

type Interface struct {
	OpenAI  openAiInterface
	Runware runwareInterface
}

func New(cfg config.Config) Interface {
	fmt.Println(cfg)
	openai := newOpenAiClient(cfg.Tokens.OpenAI())
	runware := newRunware(cfg.Tokens.Runware())
	return Interface{
		openai, runware,
	}
}
