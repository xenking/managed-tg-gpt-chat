package activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type GetRequestApprovalRequest struct {
	ChannelID    int64
	ChatID       int64
	ChatUserName string
	Request      string
}

type GetRequestApprovalResponse struct {
	MessageID int
	Message   string // TODO
	Status    domain.RequestStatus
}

func (a *Activities) GetRequestApproval(ctx context.Context, req GetRequestApprovalRequest) (GetRequestApprovalResponse, error) {
	// Retrieve the Activity information needed to asynchronously complete the Activity.
	activityInfo := activity.GetInfo(ctx)
	taskToken := activityInfo.TaskToken
	tokenID := a.TokensStorage.Store(taskToken)
	buttons := []domain.KeyboardButton{
		{
			Text:         "Approve",
			CallbackData: makeCallbackData(tokenID, domain.RequestStatusApproved),
		},
		{
			Text:         "Reject",
			CallbackData: makeCallbackData(tokenID, domain.RequestStatusRejected),
		},
	}
	content := fmt.Sprintf(`Request from <a href="tg://user?id=%d">@%s</a>:
<blockquote>%s</blockquote>`, req.ChatID, req.ChatUserName, req.Request)

	// Send a message to the chat system to request approval.
	msg, err := a.TelegramClient.SendMessageHTMLWithInlineKeyboard(ctx, req.ChannelID, content, buttons)
	if err != nil {
		return GetRequestApprovalResponse{}, err
	}

	return GetRequestApprovalResponse{
		MessageID: msg.ID,
		Message:   req.Request,
		Status:    domain.RequestStatusPending,
	}, activity.ErrResultPending
}

func makeCallbackData(tokenID uuid.UUID, action domain.RequestStatus) string {
	return tokenID.String() + ":" + string(action[:1])
}
