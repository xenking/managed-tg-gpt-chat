package domain

import (
	"context"
)

type GPTClient interface {
	Ask(ctx context.Context, msgs ...ChatMessage) (*ChatAnswer, error)
}

type ChatAnswer struct {
	Request  []ChatMessage
	Response []ChatMessage
}

type ChatMessage struct {
	Role    string
	Content string
}

const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
	ChatMessageRoleTool      = "tool"
)
