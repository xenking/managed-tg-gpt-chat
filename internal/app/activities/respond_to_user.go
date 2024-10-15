package activities

import "context"

type RespondToUserRequest struct {
	ChatID   int64
	Messages []string
}

func (a *Activities) RespondToUser(ctx context.Context, req RespondToUserRequest) error {
	if len(req.Messages) == 0 {
		return a.TelegramClient.SendMessageHTML(ctx, req.ChatID, "No response from Chat GPT")
	}
	for _, msg := range req.Messages {
		err := a.TelegramClient.SendMessageHTML(ctx, req.ChatID, msg)
		if err != nil {
			return err
		}
	}
	return nil
}
