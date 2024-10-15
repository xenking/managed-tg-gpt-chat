package adapters

import (
	"context"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
)

type TelegramClient struct {
	echotron.API
}

func NewTelegramClient(token string) *TelegramClient {
	return &TelegramClient{API: echotron.NewAPI(token)}
}

func (c *TelegramClient) SetBotCommands(ctx context.Context, commands []domain.BotCommand) error {
	var cmds []echotron.BotCommand
	for _, c := range commands {
		cmds = append(cmds, echotron.BotCommand{
			Command:     c.Command,
			Description: c.Description,
		})
	}
	_, err := c.API.SetMyCommands(ctx, nil, cmds...)
	if err != nil {
		return err
	}

	return nil
}

func (c *TelegramClient) SendMessageHTMLWithInlineKeyboard(ctx context.Context, chatID int64, message string, buttons []domain.KeyboardButton) (*domain.TelegramMessage, error) {
	var inlineKeyboard [][]echotron.InlineKeyboardButton
	for _, b := range buttons {
		inlineKeyboard = append(inlineKeyboard, []echotron.InlineKeyboardButton{
			{
				Text:         b.Text,
				CallbackData: b.CallbackData,
			},
		})
	}
	replyMarkup := echotron.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}
	opts := &echotron.MessageOptions{
		ParseMode:   echotron.HTML,
		ReplyMarkup: &replyMarkup,
	}
	res, err := c.API.SendMessage(ctx, message, chatID, opts)
	if err != nil {
		return nil, err
	}

	return &domain.TelegramMessage{
		Text:     res.Result.Text,
		ID:       res.Result.ID,
		Entities: domain.ParseTelegramMessageEntities(res.Result.Entities),
	}, nil
}

func (c *TelegramClient) SendMessage(ctx context.Context, chatID int64, message string) error {
	_, err := c.API.SendMessage(ctx, message, chatID, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *TelegramClient) SendMessageHTML(ctx context.Context, chatID int64, message string) error {
	// message = echotron.EscapeHTMLMessage(message)
	opts := &echotron.MessageOptions{
		ParseMode: echotron.HTML,
	}
	_, err := c.API.SendMessage(ctx, message, chatID, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *TelegramClient) ReplyToMessageHTML(ctx context.Context, chatID int64, messageID int, message string) error {
	// message = echotron.EscapeHTMLMessage(message)
	opts := &echotron.MessageOptions{
		ParseMode: echotron.HTML,
		ReplyParameters: echotron.ReplyParameters{
			MessageID: messageID,
			ChatID:    chatID,
		},
	}
	_, err := c.API.SendMessage(ctx, message, chatID, opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *TelegramClient) EditMessage(ctx context.Context, chatID int64, msg domain.TelegramMessage) error {
	opts := &echotron.MessageTextOptions{
		Entities: msg.Entities,
	}
	_, err := c.API.EditMessageText(ctx, msg.Text, echotron.NewMessageID(chatID, msg.ID), opts)
	if err != nil {
		return err
	}

	return nil
}

func (c *TelegramClient) EditMessageHTML(ctx context.Context, chatID int64, msg domain.TelegramMessage) error {
	// msg.Text = echotron.EscapeHTMLMessage(msg.Text)
	opts := &echotron.MessageTextOptions{
		ParseMode: echotron.HTML,
		Entities:  msg.Entities,
	}
	_, err := c.API.EditMessageText(ctx, msg.Text, echotron.NewMessageID(chatID, msg.ID), opts)
	if err != nil {
		return err
	}

	return nil
}
