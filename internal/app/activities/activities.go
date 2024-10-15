package activities

import (
	"go.temporal.io/sdk/client"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type Activities struct {
	Client         client.Client
	TelegramClient domain.TelegramClient
	GPTClient      domain.GPTClient
	HTMlConverter  domain.MarkdownHTMLConverter
	TokensStorage  domain.ActivitiesTokenStorage
}

func New(cli client.Client, tgCli domain.TelegramClient, gptClient domain.GPTClient,
	storage domain.ActivitiesTokenStorage,
	converter domain.MarkdownHTMLConverter,
) *Activities {
	return &Activities{
		Client:         cli,
		TelegramClient: tgCli,
		GPTClient:      gptClient,
		TokensStorage:  storage,
		HTMlConverter:  converter,
	}
}
