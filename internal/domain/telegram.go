package domain

import (
	"context"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
	"github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

type TelegramClient interface {
	SendMessage(ctx context.Context, chatID int64, message string) error
	SetBotCommands(ctx context.Context, commands []BotCommand) error
	EditMessage(ctx context.Context, chatID int64, msg TelegramMessage) error

	SendMessageHTML(ctx context.Context, chatID int64, message string) error
	SendMessageHTMLWithInlineKeyboard(ctx context.Context, chatID int64, message string, buttons []KeyboardButton) (*TelegramMessage, error)
	EditMessageHTML(ctx context.Context, chatID int64, msg TelegramMessage) error
	ReplyToMessageHTML(ctx context.Context, chatID int64, messageID int, message string) error
}

type TelegramRoute interface {
	tgrouter.Route
}

type TelegramMessage struct {
	Text     string
	ID       int
	Entities []echotron.MessageEntity
}

func ParseTelegramMessageEntities(entities []*echotron.MessageEntity) []echotron.MessageEntity {
	var res []echotron.MessageEntity
	for _, e := range entities {
		if e != nil {
			res = append(res, *e)
		}
	}
	return res
}

type KeyboardButton struct {
	Text         string
	CallbackData string
}

type BotCommand struct {
	Command     string
	Description string
}
