package ports

import (
	"context"
	"log/slog"

	"github.com/xenking/managed-tg-gpt-chat/pkg/echotron"
	"github.com/xenking/managed-tg-gpt-chat/pkg/tgrouter"
)

type BotServer struct {
	*echotron.Dispatcher
	router *tgrouter.Router
	logger *slog.Logger
}

func NewBotServer(token string, allowedChats []int64, logger *slog.Logger) *BotServer {
	b := &BotServer{
		logger: logger.With(slog.String("component", "BotServer")),
	}
	dsp := echotron.NewDispatcher(token, b.newBotSession(allowedChats))
	b.Dispatcher = dsp

	api := echotron.NewAPI(token)
	b.router = tgrouter.NewRouter(api,
		tgrouter.WithNotFoundHandler(tgrouter.HandlerFunc(b.notFoundHandler)),
		tgrouter.WithErrorHandler(b.errorHandler),
		tgrouter.WithRecoverHandler(b.panicHandler),
	)
	return b
}

func (b *BotServer) Mount(routes ...tgrouter.Route) *BotServer {
	b.router.Mount(routes...)
	return b
}

func (b *BotServer) Start(ctx context.Context) {
	go func() {
		err := b.Dispatcher.PollOptions(ctx, true, echotron.UpdateOptions{
			AllowedUpdates: []echotron.UpdateType{
				echotron.UpdateTypeMessage,
				echotron.UpdateTypeCallbackQuery,
			},
			Timeout: 120, // 2 minutes
		})
		if err != nil {
			b.logger.ErrorContext(ctx, "dispatcher", slog.Any("error", err))
		}
	}()
	go b.Dispatcher.ListenUpdates(ctx)
}

func (b *BotServer) newBotSession(allowedChatIDs []int64) func(chatID int64) echotron.SessionHandler {
	allowedChats := map[int64]struct{}{}
	for _, chat := range allowedChatIDs {
		allowedChats[chat] = struct{}{}
	}
	return func(chatID int64) echotron.SessionHandler {
		if _, ok := allowedChats[chatID]; !ok {
			return echotron.NoopSessionHandler
		}
		return b.router
	}
}

func (b *BotServer) errorHandler(ctx context.Context, u *tgrouter.Update, err error) {
	b.logger.ErrorContext(ctx, "error handler", slog.Any("error", err), slog.Any("update", u))
}

func (b *BotServer) notFoundHandler(ctx context.Context, u *tgrouter.Update) error {
	b.logger.WarnContext(ctx, "route not found", slog.Any("update", u))
	return nil
}

func (b *BotServer) panicHandler(u *tgrouter.Update, err error) {
	b.logger.Error("panic", slog.Any("error", err), slog.Any("update", u))
}
