package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jyouturer/gmail-ai/internal"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
)

type Message struct {
	ID      string
	Subject string
	From    string
	To      string
	Body    string
	Payload []byte
}

// Define a type for message handler functions
type MessageHandlerFunc func(ctx context.Context, msg Message) error

// Define an interface for message services
type MessageService interface {
	GetMessageIds(userId string, startHistoryId uint64) (uint64, []string, error)
	GetMessage(userId string, id string) (Message, error)
	GetMessages(userId string, ids []string) ([]Message, error)
}

// Define a struct to represent a message provider
type MessageProvider struct {
	service MessageService
}

func NewMessageProvider(service MessageService) *MessageProvider {
	return &MessageProvider{
		service: service,
	}
}

func (ep *MessageProvider) PollAndProcess(ctx context.Context, pollHistory PollHistory, handlers []MessageHandlerFunc) error {
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
	logging.Logger.Debug("last history ID", zap.Uint64("historyId", lastHistoryId))
	// Retrieve the list of histories
	lastHistoryId, ids, err := ep.service.GetMessageIds("me", lastHistoryId)
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
	logging.Logger.Debug("message ids", zap.Any("message ids", ids))
	// Create a wait group for email handler functions
	var wg sync.WaitGroup

	// Process each message in the history item
	for _, id := range ids {
		// Check if the message has already been processed
		if processedMessages.Contains(id) {
			continue
		}
		// Retrieve the message
		logging.Logger.Debug("retrieving message", zap.String("message", id))
		m, err := ep.service.GetMessage("me", id)
		if err != nil {
			return fmt.Errorf("unable to get message: %v", err)
		}
		// Process the message content with each handler function to determine if it meets the criteria
		for _, handler := range handlers {
			wg.Add(1)
			go handleMessage(ctxTimeout, handler, m, &wg)
		}

		// Wait for all handler functions to complete or for a context timeout
		if err = waitHandlers(ctxTimeout, &wg, processedMessages, id); err != nil {
			return err
		}
	}

	return nil
}

func handleMessage(ctx context.Context, handler MessageHandlerFunc, msg Message, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := handler(ctx, msg); err != nil {
		logging.Logger.Error("error processing message", zap.String("message", msg.ID), zap.Error(err))
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
