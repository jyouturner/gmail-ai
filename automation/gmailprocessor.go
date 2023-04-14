package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jyouturer/gmail-ai/internal"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
)

// Define a type for email handler functions
type EmailHandlerFunc func(ctx context.Context, email *gmail.Message) error

// Define an interface for email services
type EmailService interface {
	GetHistoryList(userId string, startHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error)
	GetMessage(userId string, id string) (*gmail.Message, error)
}

// Define a struct to represent an email provider
type EmailProvider struct {
	service EmailService
}

func NewEmailProvider(service EmailService) *EmailProvider {
	return &EmailProvider{
		service: service,
	}
}

// PollAndProcess retrieves new emails and processes them
func (ep *EmailProvider) PollAndProcess(ctx context.Context, pollHistory PollHistory, handlers []EmailHandlerFunc) error {
	// Create a context with a timeout of 60 seconds
	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Create a processed messages buffer
	processedMessages := internal.NewRingBuffer(100)

	// Read the last history ID from the file
	lastHistoryId, err := pollHistory.ReadHistory()
	if err != nil {
		return fmt.Errorf("unable to read last historyId from file: %v", err)
	}

	// Retrieve the list of histories
	lastHistoryId, histories, err := ep.service.GetHistoryList("me", lastHistoryId)
	if err != nil {
		return fmt.Errorf("unable to get histories: %v", err)
	}

	// Defer writing the last history ID to the file
	defer func() {
		err = pollHistory.WriteHistory(lastHistoryId)
		if err != nil {
			logging.Logger.Error("error saving history", zap.Any("history", pollHistory), zap.Error(err))
		}
	}()

	// Create a wait group for email handler functions
	var wg sync.WaitGroup

	// Process each history item
	for _, h := range histories.History {
		lastHistoryId = h.Id

		// Process each message in the history item
		for _, m := range h.MessagesAdded {
			// Check if the message has already been processed
			messageID := m.Message.Id
			if processedMessages.Contains(messageID) {
				continue
			}

			// Retrieve the message
			msg, err := ep.service.GetMessage("me", messageID)
			if err != nil {
				continue
			}

			// Process the email content with each handler function to determine if it meets the criteria
			for _, handler := range handlers {
				wg.Add(1)
				go handleEmail(ctxTimeout, handler, messageID, msg, &wg)
			}

			// Wait for all email handler functions to complete or for a context timeout
			if err = waitHandlers(ctxTimeout, &wg, processedMessages, messageID); err != nil {
				return err
			}
		}
	}

	return nil
}

func handleEmail(ctx context.Context, handler EmailHandlerFunc, messageID string, msg *gmail.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := handler(ctx, msg); err != nil {
		logging.Logger.Error("error processing email", zap.String("message", messageID), zap.Error(err))
	}
}

func waitHandlers(ctx context.Context, wg *sync.WaitGroup, processedMessages *internal.RingBuffer, messageID string) error {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		wg.Wait()
	}()

	select {
	case <-doneCh:
		// All email handler functions have completed
		// Mark the message as processed in the ring buffer
		processedMessages.Put(messageID)
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Retrieve a message by ID
func (ep *EmailProvider) GetMessage(userId string, id string) (*gmail.Message, error) {
	return ep.service.GetMessage(userId, id)
}

// Retrieve a list of histories
func (ep *EmailProvider) GetHistoryList(userId string, startHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error) {
	return ep.service.GetHistoryList(userId, startHistoryId)
}
