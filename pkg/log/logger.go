package log

import (
	"context"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
)

type TelegramWriter struct {
	api    echotron.API
	chatID int64
}

func NewTelegramWriter(token string, chatID int64) *TelegramWriter {
	return &TelegramWriter{
		api:    echotron.NewAPI(token),
		chatID: chatID,
	}
}

func (w TelegramWriter) Write(p []byte) (n int, err error) {
	_, err = w.api.SendMessage(context.Background(), string(p), w.chatID, nil)
	return len(p), err
}
