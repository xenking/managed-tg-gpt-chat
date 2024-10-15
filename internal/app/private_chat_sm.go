package app

import (
	"context"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/xenking/managed-tg-gpt-chat/internal/app/activities"
	"github.com/xenking/managed-tg-gpt-chat/internal/app/workflows"
	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type ChatMessage struct {
	Message   string
	MessageID int
}

// PrivateChatStateMachine maintains current state and shared dependencies.
type PrivateChatStateMachine struct {
	*Service
	domain.StateMachine[*ChatMessage]
	chatID   int64
	userName string

	WorkflowActivityID string
	WorkflowID         string
	WorkflowRunID      string
}

func (s *Service) NewPrivateChatStateMachine(chatID int64, userName string) *PrivateChatStateMachine {
	sm := &PrivateChatStateMachine{
		Service:            s,
		chatID:             chatID,
		userName:           userName,
		WorkflowActivityID: uuid.New().String(),
	}
	sm.StateMachine = domain.NewStateMachine[*ChatMessage](sm.StateNoop)
	return sm
}

func (sm *PrivateChatStateMachine) StateNoop(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	return sm.StateNoop, nil
}

func (sm *PrivateChatStateMachine) StateNewDialog(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	err := sm.telegram.SendMessage(ctx, sm.chatID, "Please enter your request")
	if err != nil {
		return nil, err
	}
	return sm.StateListenRequest, nil
}

func (sm *PrivateChatStateMachine) StateListenRequest(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	workflow, err := sm.temporal.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		TaskQueue: domain.ChatRequestsQueue,
	}, workflows.ChatGTPSession, workflows.ChatGPTSessionInput{
		ChatID:             sm.chatID,
		ChatUserName:       sm.userName,
		Request:            msg.Message,
		WorkflowActivityID: sm.WorkflowActivityID,
		AuditLogChannelID:  sm.cfg.AuditLogChannelID,
	})
	if err != nil {
		return nil, err
	}
	sm.WorkflowID = workflow.GetID()
	sm.WorkflowRunID = workflow.GetRunID()
	err = sm.telegram.SendMessage(ctx, sm.chatID, "Waiting for approval...")
	return sm.StateWaitForApprove, err
}

func (sm *PrivateChatStateMachine) StateWaitForApprove(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	err := sm.telegram.SendMessage(ctx, sm.chatID, "Waiting for approval...\nYou can cancel the request by /cancel")
	return sm.StateWaitForApprove, err
}

func (sm *PrivateChatStateMachine) StateContinueConversation(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	// TODO: Continue conversation in new workflow (or old one)
	return sm.StateListenRequest, nil
}

func (sm *PrivateChatStateMachine) StateCancel(ctx context.Context, msg *ChatMessage) (domain.StateFunc[*ChatMessage], error) {
	defer sm.Reset()
	err := sm.temporal.CompleteActivityByID(ctx,
		"default",
		sm.WorkflowID, sm.WorkflowRunID,
		sm.WorkflowActivityID,
		activities.GetRequestApprovalResponse{
			MessageID: msg.MessageID,
			Message:   "Request canceled",
			Status:    domain.RequestStatusCanceled,
		},
		nil)
	if err != nil {
		return nil, err
	}
	err = sm.telegram.SendMessage(ctx, sm.chatID, "Request canceled")
	return sm.StateNoop, err
}

func (sm *PrivateChatStateMachine) Reset() {
	sm.WorkflowRunID = ""
	sm.WorkflowID = ""
	sm.WorkflowActivityID = uuid.New().String()
}
