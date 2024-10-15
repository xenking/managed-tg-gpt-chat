package activities

import (
	"context"
)

type CommentRequestWithResponse struct {
	GroupID   int64
	MessageID int
	Request   string
	Responses []string
}

func (a *Activities) CommentRequestWithResponse(ctx context.Context, req CommentRequestWithResponse) error {
	for _, msg := range req.Responses {
		err := a.TelegramClient.ReplyToMessageHTML(ctx, req.GroupID, req.MessageID, msg)
		if err != nil {
			return err
		}
	}

	return nil
}
