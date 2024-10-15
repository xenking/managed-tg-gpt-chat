package activities

import (
	"context"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type SendChatGPTRequestRequest struct {
	Request string
}

type SendChatGPTRequestResponse struct {
	Responses []domain.ChatMessage
}

func (a *Activities) SendChatGPTRequest(ctx context.Context, req SendChatGPTRequestRequest) (SendChatGPTRequestResponse, error) {
	answer, err := a.GPTClient.Ask(ctx, domain.ChatMessage{Role: domain.ChatMessageRoleUser, Content: req.Request})
	if err != nil {
		return SendChatGPTRequestResponse{}, err
	}

	return SendChatGPTRequestResponse{
		Responses: answer.Response,
	}, nil
}
