// main.go
package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/xenking/managed-tg-gpt-chat/internal/adapters"
	"github.com/xenking/managed-tg-gpt-chat/internal/app"
	"github.com/xenking/managed-tg-gpt-chat/internal/app/activities"
	"github.com/xenking/managed-tg-gpt-chat/internal/app/workflows"
	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
	"github.com/xenking/managed-tg-gpt-chat/internal/ports"
	"github.com/xenking/managed-tg-gpt-chat/pkg/log"
)

var (
	errorLogChatID         = int64(203335723)      // ERROR_LOG_CHAT_ID
	auditLogChannelID      = int64(-1002404642431) // AUDIT_LOG_CHANNEL_ID
	auditLogChannelGroupID = int64(-1002364910005) // AUDIT_LOG_CHANNEL_GROUP_ID
)

var allowedChatIDs = []int64{
	203335723,  // @xenking
	707549989,  // @tishchenkoanna
	2072665059, // @andrewtishchenko
	auditLogChannelID,
	auditLogChannelGroupID,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	gptKey := os.Getenv("GPT_API_KEY")
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")

	logger := slog.New(slog.NewTextHandler(
		io.MultiWriter(
			os.Stdout,
			log.NewTelegramWriter(tgToken, errorLogChatID),
		), nil),
	)
	slog.SetDefault(logger)

	// Set up the Temporal client.
	temporalClient, err := client.DialContext(ctx, client.Options{
		HostPort: temporalAddr,
	})
	if err != nil {
		logger.Error("Unable to connect to Temporal server", slog.Any("error", err))
		panic(err)
	}
	defer temporalClient.Close()

	tgClient := adapters.NewTelegramClient(tgToken)

	gptClient := adapters.NewGPTClient(gptKey)
	activityTokenStorage := adapters.NewInMemoryTokenStorage()
	markdownHTmlConverter := adapters.NewMarkdownHTMLConverter()

	act := activities.New(temporalClient, tgClient, gptClient, activityTokenStorage, markdownHTmlConverter)

	err = StartWorker(ctx, temporalClient, act)
	if err != nil {
		logger.Error("start worker", slog.Any("error", err))
		panic(err)
	}

	cfg := app.Config{
		WhitelistedUsers:       allowedChatIDs,
		AuditLogChannelID:      auditLogChannelID,
		AuditLogChannelGroupID: auditLogChannelGroupID,
	}

	service := app.NewService(tgClient, temporalClient, activityTokenStorage, logger, cfg)

	server := ports.NewBotServer(tgToken, allowedChatIDs, logger).Mount(
		service.PrivateChatRoutes(),
		service.GroupRoutes(),
	)
	go server.Start(ctx)

	<-ctx.Done()
}

func StartWorker(ctx context.Context, cli client.Client, a *activities.Activities) error {
	// Set up the Temporal worker.
	w := worker.New(cli, domain.ChatRequestsQueue, worker.Options{})

	w.RegisterWorkflow(workflows.ChatGTPSession)
	w.RegisterActivity(a.GetRequestApproval)
	w.RegisterActivity(a.SendChatGPTRequest)
	w.RegisterActivity(a.RejectChatRequest)
	w.RegisterActivity(a.RespondToUser)
	w.RegisterActivity(a.ConvertToHTML)
	w.RegisterActivity(a.CommentRequestWithResponse)

	err := w.Start()
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		w.Stop()
	}()
	return nil
}
