package adapters

import (
	"context"

	"github.com/sashabaranov/go-openai"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type GPTClient struct {
	*openai.Client
}

func NewGPTClient(token string) *GPTClient {
	return &GPTClient{
		Client: openai.NewClient(token),
	}
}

var systemPrompt = openai.ChatCompletionMessage{
	Role: openai.ChatMessageRoleSystem,
	Content: `You are a senior front-end developer who helps the junior dev. 
Explain all your actions. You can answer in Russian or English. 
STRICT RULE: You can use only telegram html style formatting.`,
}

func (c *GPTClient) Ask(ctx context.Context, msgs ...domain.ChatMessage) (*domain.ChatAnswer, error) {
	requestMessages := []openai.ChatCompletionMessage{systemPrompt}
	for _, msg := range msgs {
		requestMessages = append(requestMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	resp, err := c.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: requestMessages,
		},
	)
	if err != nil {
		return nil, err
	}
	var response []domain.ChatMessage
	for _, msg := range resp.Choices {
		response = append(response, domain.ChatMessage{
			Role:    msg.Message.Role,
			Content: msg.Message.Content,
		})
	}

	return &domain.ChatAnswer{
		Request:  msgs,
		Response: response,
	}, nil
}
