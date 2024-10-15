package activities

import (
	"context"
	"errors"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type ConvertToHTMLRequest struct {
	ChatResponse []domain.ChatMessage
}

type ConvertToHTMLResponse struct {
	HTMLContent string
}

func (a *Activities) ConvertToHTML(ctx context.Context, req ConvertToHTMLRequest) (ConvertToHTMLResponse, error) {
	var lastMessage string
	for i := len(req.ChatResponse) - 1; i >= 0; i-- {
		if req.ChatResponse[i].Role == domain.ChatMessageRoleAssistant {
			lastMessage = req.ChatResponse[i].Content
			break
		}
	}
	if lastMessage == "" {
		return ConvertToHTMLResponse{}, errors.New("no assistant message found")
	}

	htmlContent, err := a.HTMlConverter.ConvertToHTML(ctx, lastMessage)
	if err != nil {
		return ConvertToHTMLResponse{}, err
	}

	return ConvertToHTMLResponse{
		HTMLContent: htmlContent,
	}, nil
}
