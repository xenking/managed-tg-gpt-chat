package workflows

import (
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/xenking/managed-tg-gpt-chat/internal/app/activities"
	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

const GetGroupMessageSignal = "group-message-id-signal"

type GetGroupMessageInput struct {
	MessageID int
	GroupID   int64
}

type ChatGPTSessionInput struct {
	WorkflowActivityID string
	AuditLogChannelID  int64

	ChatID       int64
	ChatUserName string
	Request      string
}

type ChatGPTSessionOutput struct {
	Status   domain.RequestStatus
	Response string
}

// ChatGTPSession is a Temporal workflow
// that orchestrates the approval of a chat GPT request
// activities.GetRequestApproval
// switch based on response
// activities.RequestToChatGPT or activities.RejectChatRequest
// activities.ConvertToHTML
// activities.RespondToUser
// wait for new message in group
// activities.CommentRequestWithResponse
// TODO: child workflow (with approval) for chat continuation
func ChatGTPSession(ctx workflow.Context, input ChatGPTSessionInput) (ChatGPTSessionOutput, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout:    time.Hour,
		ScheduleToCloseTimeout: 10 * time.Minute,
		ActivityID:             input.WorkflowActivityID,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 5 * time.Second,
			MaximumAttempts: 10,
		},
	})

	var approvalResp activities.GetRequestApprovalResponse
	err := workflow.ExecuteActivity(ctx, a.GetRequestApproval, activities.GetRequestApprovalRequest{
		ChannelID:    input.AuditLogChannelID,
		ChatID:       input.ChatID,
		ChatUserName: input.ChatUserName,
		Request:      input.Request,
	}).Get(ctx, &approvalResp)
	if err != nil {
		return ChatGPTSessionOutput{}, err
	}

	switch approvalResp.Status {
	case domain.RequestStatusRejected, domain.RequestStatusCanceled:
		err = workflow.ExecuteActivity(ctx, a.RejectChatRequest, activities.RejectChatRequestRequest{
			ChatID:        input.ChatID,
			RejectMessage: approvalResp.Message,
		}).Get(ctx, nil)
		return ChatGPTSessionOutput{
			Status:   approvalResp.Status,
			Response: fmt.Sprintf("Request %s", approvalResp.Status),
		}, err
	}

	var chatResp activities.SendChatGPTRequestResponse
	err = workflow.ExecuteActivity(ctx, a.SendChatGPTRequest, activities.SendChatGPTRequestRequest{
		Request: input.Request,
	}).Get(ctx, &chatResp)
	if err != nil {
		return ChatGPTSessionOutput{}, err
	}

	var htmlResp activities.ConvertToHTMLResponse
	err = workflow.ExecuteActivity(ctx, a.ConvertToHTML, activities.ConvertToHTMLRequest{
		ChatResponse: chatResp.Responses,
	}).Get(ctx, &htmlResp)
	if err != nil {
		return ChatGPTSessionOutput{}, err
	}

	messages := splitMaxLimitMessages(htmlResp.HTMLContent)

	err = workflow.ExecuteActivity(ctx, a.RespondToUser, activities.RespondToUserRequest{
		ChatID:   input.ChatID,
		Messages: messages,
	}).Get(ctx, nil)
	if err != nil {
		return ChatGPTSessionOutput{}, err
	}

	// Block until we receive a new message in group
	var groupMessageInput GetGroupMessageInput
	workflow.GetSignalChannel(ctx, GetGroupMessageSignal).Receive(ctx, &groupMessageInput)

	err = workflow.ExecuteActivity(ctx, a.CommentRequestWithResponse, activities.CommentRequestWithResponse{
		GroupID:   groupMessageInput.GroupID,
		MessageID: groupMessageInput.MessageID,
		Request:   input.Request,
		Responses: messages,
	}).Get(ctx, nil)
	if err != nil {
		return ChatGPTSessionOutput{}, err
	}

	return ChatGPTSessionOutput{
		Status:   domain.RequestStatusApproved,
		Response: fmt.Sprintf("Responses: %s", chatResp.Responses),
	}, nil
}

const MaxTelegramMessageSize = 4096 - 128

func splitMaxLimitMessages(text string) []string {
	var parts []string
	var currentPart strings.Builder
	var openTagStack []string

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if currentPart.Len()+len(line) > MaxTelegramMessageSize {
			parts = append(parts, closeOpenTags(&currentPart, openTagStack))

			// Prepare the next part
			line = strings.TrimSpace(line)

			// Re-open the last tag if there were any
			for _, tag := range openTagStack {
				currentPart.WriteString(fmt.Sprintf("<%s>", tag))
			}
		}

		currentPart.WriteString(line + "\n")

		// Tag handling within each line
		for i := 0; i < len(line); {
			startIdx := strings.Index(line[i:], "<")
			if startIdx == -1 {
				break
			}
			startIdx += i
			endIdx := strings.Index(line[startIdx:], ">")
			if endIdx == -1 {
				break
			}
			endIdx += startIdx

			tagContent := line[startIdx+1 : endIdx]

			if strings.HasPrefix(tagContent, "/") {
				// It's a closing tag
				tagName := strings.Fields(strings.TrimSpace(tagContent[1:]))[0]

				if len(openTagStack) > 0 && openTagStack[len(openTagStack)-1] == tagName {
					openTagStack = openTagStack[:len(openTagStack)-1]
				}
			} else {
				// It's an opening tag
				tagName := strings.Fields(strings.TrimSpace(tagContent))[0]
				if !strings.HasSuffix(tagContent, "/") && !strings.HasPrefix(tagContent, "!--") && !strings.HasPrefix(tagContent, "?") {
					openTagStack = append(openTagStack, tagName)
				}
			}
			i = endIdx + 1
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, closeOpenTags(&currentPart, openTagStack))
	}

	return parts
}

func closeOpenTags(currentPart *strings.Builder, openTagStack []string) string {
	for j := len(openTagStack) - 1; j >= 0; j-- {
		currentPart.WriteString(fmt.Sprintf("</%s>", openTagStack[j]))
	}
	part := currentPart.String()
	currentPart.Reset()
	return part
}
