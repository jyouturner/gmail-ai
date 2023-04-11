package automation

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/jyouturer/gmail-ai/internal"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
)

// Define a type for email handler functions
type EmailHandlerFunc func(ctx context.Context, email *gmail.Message) error

// ProcessNewEmails retrieves new emails and processes them
func ProcessNewEmails(ctx context.Context, gmailService *gmail.Service, historyFile string, handlers []EmailHandlerFunc) error {

	processedMessages := internal.NewRingBuffer(100)

	// Read the last historyId from the file
	lastHistoryId, err := readHistoryId(historyFile)
	if err != nil {
		return fmt.Errorf("unable to read last historyId from file: %v", err)
	}

	lastHistoryId, histories, err := integration.GetHistoryList(gmailService, "me", lastHistoryId)
	if err != nil {
		return fmt.Errorf("unable to get histories %v", err)
	}
	logging.Logger.Debug("History:", zap.Int("count", len(histories.History)))
	// Create a WaitGroup to wait for all email handler functions to complete
	var wg sync.WaitGroup
	// Process each message returned by the history API
	for _, h := range histories.History {
		logging.Logger.Debug("History", zap.Uint64("id", h.Id))
		lastHistoryId = h.Id
		for _, m := range h.MessagesAdded {
			// Check if the message has already been processed
			messageID := m.Message.Id
			if processedMessages.Contains(messageID) {
				logging.Logger.Info("Message has already been processed, skipping...\n", zap.String("message", messageID))
				continue
			}

			logging.Logger.Info("Message", zap.String("message", messageID))
			// Retrieve only the message headers to limit the size of the response
			msg, err := gmailService.Users.Messages.Get("me", messageID).Format("full").Do()
			if err != nil {
				logging.Logger.Error("unable to retrieve message", zap.String("message", messageID), zap.Error(err))
				continue
			}
			snippet := msg.Snippet
			logging.Logger.Info(snippet)

			// Process the email content with each handler function to determine if it meets the criteria
			for _, handler := range handlers {
				wg.Add(1)
				go func(h EmailHandlerFunc) {
					defer wg.Done()
					err := h(ctx, msg)
					if err != nil {
						logging.Logger.Error("error processing email", zap.String("message", messageID), zap.Error(err))
					}
				}(handler)
			}

			// Wait for all email handler functions to complete or for a context timeout
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
		}
	}

	// Write the updated lastHistoryId to the file
	err = writeHistoryId(historyFile, lastHistoryId)
	if err != nil {
		logging.Logger.Error("error saving history file", zap.String("historyFile", historyFile), zap.Error(err))
		return fmt.Errorf("unable to write last historyId to file: %v", err)
	}
	return nil
}

// Read the last historyId from the file
func readHistoryId(filename string) (uint64, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// If the file does not exist, return 0 and no error
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	historyId, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return historyId, nil
}

// Write the last historyId to the file
func writeHistoryId(filename string, historyId uint64) error {
	return ioutil.WriteFile(filename, []byte(strconv.FormatUint(historyId, 10)), 0644)
}
