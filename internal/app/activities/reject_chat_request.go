package activities

import (
	"context"
)

type RejectChatRequestRequest struct {
	ChatID        int64
	RejectMessage string
}

func (a *Activities) RejectChatRequest(ctx context.Context, req RejectChatRequestRequest) error {
	return a.TelegramClient.SendMessage(ctx, req.ChatID, req.RejectMessage)
}
