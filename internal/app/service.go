package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/xenking/managed-tg-gpt-chat/internal/app/activities"
	"github.com/xenking/managed-tg-gpt-chat/internal/app/workflows"
	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
	"github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

type Config struct {
	WhitelistedUsers       []int64
	AuditLogChannelID      int64
	AuditLogChannelGroupID int64
}

type Service struct {
	telegram     domain.TelegramClient
	temporal     domain.TemporalClient
	tokenStorage domain.ActivitiesTokenStorage
	logger       *slog.Logger
	cfg          Config
	dialogs      map[int64]*PrivateChatStateMachine
	mu           sync.RWMutex
}

func NewService(
	telegramClient domain.TelegramClient,
	temporalClient domain.TemporalClient,
	tokenStorage domain.ActivitiesTokenStorage,
	logger *slog.Logger,
	cfg Config,
) *Service {
	return &Service{
		telegram:     telegramClient,
		temporal:     temporalClient,
		tokenStorage: tokenStorage,
		dialogs:      make(map[int64]*PrivateChatStateMachine),
		logger:       logger,
		cfg:          cfg,
	}
}

func (s *Service) PrivateChatRoutes() domain.TelegramRoute {
	return tgrouter.NewGroup(tgrouter.IsPrivate(),
		tgrouter.NewCommandRoute("/start", nil, tgrouter.HandlerFunc(s.handleStartCommand)),
		tgrouter.NewCommandRoute("/ask", nil, tgrouter.HandlerFunc(s.handleStateMachineCreate)),
		tgrouter.NewCommandRoute("/cancel", nil, tgrouter.HandlerFunc(s.handleStateMachineCancel)),
		tgrouter.NewMessageRoute(nil, tgrouter.HandlerFunc(s.handleStateMachine)),
	)
}

func (s *Service) GroupRoutes() domain.TelegramRoute {
	return tgrouter.NewGroup(tgrouter.Or(tgrouter.IsChannel(), tgrouter.IsSuperGroup()),
		tgrouter.NewRoute(tgrouter.IsCallbackQuery(), tgrouter.HandlerFunc(
			func(ctx context.Context, u *tgrouter.Update) error {
				return s.handleCompleteActivity(ctx, u.CallbackQuery)
			}),
		),
		tgrouter.NewRoute(tgrouter.And(tgrouter.IsSuperGroup(), tgrouter.IsForwardOriginType("channel")), tgrouter.HandlerFunc(
			func(ctx context.Context, u *tgrouter.Update) error {
				return s.handleForwarderGroupMessage(ctx, u)
			}),
		),
		tgrouter.NewRoute(tgrouter.IsChannelPost(), tgrouter.HandlerFunc(
			func(ctx context.Context, u *tgrouter.Update) error {
				log.Println("Channel post", u.Message.Text)
				return nil
			}),
		),
	)
}

func (s *Service) handleCompleteActivity(ctx context.Context, q *echotron.CallbackQuery) error {
	activityToken, status, err := s.parseCallbackData(q.Data)
	if err != nil {
		_ = s.telegram.EditMessageHTML(ctx, q.Message.Chat.ID, domain.TelegramMessage{
			Text:     fmt.Sprintf("%s\n\nStatus: <b>%s</b>", q.Message.Text, domain.RequestStatusCanceled),
			ID:       q.Message.ID,
			Entities: domain.ParseTelegramMessageEntities(q.Message.Entities),
		})
		return err
	}
	err = s.temporal.CompleteActivity(ctx, activityToken, activities.GetRequestApprovalResponse{
		MessageID: q.Message.ID,
		Message:   fmt.Sprintf("Request %s", status),
		Status:    status,
	}, nil)
	if err != nil {
		return err
	}
	err = s.telegram.EditMessageHTML(ctx, q.Message.Chat.ID, domain.TelegramMessage{
		Text:     fmt.Sprintf("%s\n\nStatus: <b>%s</b>", q.Message.Text, status),
		ID:       q.Message.ID,
		Entities: domain.ParseTelegramMessageEntities(q.Message.Entities),
	})
	// Reset the state machine if the request was rejected
	if status == domain.RequestStatusRejected {
		chatID := getUserFromMessageEntities(q.Message)
		sm := s.getDialogSM(chatID)
		sm.Reset()
		sm.Set(sm.StateNoop)
	}
	return err
}

func (s *Service) parseCallbackData(data string) ([]byte, domain.RequestStatus, error) {
	parts := strings.Split(data, ":")
	uid := uuid.MustParse(parts[0])
	activityToken, ok := s.tokenStorage.Pop(uid)
	if !ok {
		return nil, domain.RequestStatusRejected, fmt.Errorf("activity token not found for callback ID %s", parts[0])
	}
	switch parts[1] {
	case "a":
		return activityToken, domain.RequestStatusApproved, nil
	case "r":
		return activityToken, domain.RequestStatusRejected, nil
	}
	return nil, domain.RequestStatusRejected, fmt.Errorf("unknown callback data %s", parts[1])
}

var sendCommandsOnce sync.Once

func (s *Service) handleStartCommand(ctx context.Context, u *tgrouter.Update) error {
	err := s.telegram.SendMessage(ctx,
		u.ChatID(),
		"Hello! I'm a GPT-4-mini chatbot. Send me a message and I'll respond with a generated text.",
	)
	if err != nil {
		return err
	}
	sendCommandsOnce.Do(func() {
		cmds := []domain.BotCommand{
			{
				Command:     "/ask",
				Description: "New question to ChatGPT",
			},
			{
				Command:     "/cancel",
				Description: "Cancel the current ask request",
			},
		}
		err = s.telegram.SetBotCommands(ctx, cmds)
	})
	return err
}

func (s *Service) handleStateMachineCreate(ctx context.Context, u *tgrouter.Update) error {
	sm := s.getDialogSM(u.ChatID())
	if sm == nil {
		sm = s.NewPrivateChatStateMachine(u.ChatID(), u.Message.From.UserName)
		s.mu.Lock()
		s.dialogs[u.ChatID()] = sm
		s.mu.Unlock()
	}
	sm.Set(sm.StateNewDialog)
	err := sm.Execute(ctx, &ChatMessage{Message: u.Message.Text, MessageID: u.Message.ID})
	return err
}

func (s *Service) handleStateMachine(ctx context.Context, u *tgrouter.Update) error {
	sm := s.getDialogSM(u.ChatID())
	if sm == nil {
		return fmt.Errorf("active state machine not found for chat %d", u.ChatID())
	}
	err := sm.Execute(ctx, &ChatMessage{Message: u.Message.Text, MessageID: u.Message.ID})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) handleStateMachineCancel(ctx context.Context, u *tgrouter.Update) error {
	sm := s.getDialogSM(u.ChatID())
	if sm == nil {
		// No active state machine - nothing to cancel
		return nil
	}
	sm.Set(sm.StateCancel)
	err := sm.Execute(ctx, &ChatMessage{Message: u.Message.Text, MessageID: u.Message.ID})
	return err
}

func (s *Service) getDialogSM(chatID int64) *PrivateChatStateMachine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dialogs[chatID]
}

func (s *Service) handleForwarderGroupMessage(ctx context.Context, u *tgrouter.Update) error {
	chatID := getUserFromMessageEntities(u.Update.Message)
	sm := s.getDialogSM(chatID)
	if sm == nil {
		return fmt.Errorf("active state machine not found for chat %d", chatID)
	}
	err := s.temporal.SignalWorkflow(ctx, sm.WorkflowID, sm.WorkflowRunID, workflows.GetGroupMessageSignal,
		workflows.GetGroupMessageInput{
			MessageID: u.Message.ID,
			GroupID:   u.ChatID(),
		})
	if err != nil {
		return err
	}
	return nil
}

func getUserFromMessageEntities(msg *echotron.Message) int64 {
	if len(msg.Entities) > 0 && msg.Entities[0].Type == "text_mention" {
		return msg.Entities[0].User.ID
	}
	return -1
}
