package domain

import "go.temporal.io/sdk/client"

const ChatRequestsQueue = "chat-requests-queue"

type TemporalClient interface {
	client.Client
}
