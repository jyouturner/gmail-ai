package automation

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"golang.org/x/time/rate"
	"google.golang.org/api/gmail/v1"
)

type RateLimiter struct {
	limiter  *rate.Limiter
	requests int
}

func NewRateLimiter(r rate.Limit, b int, requests int) *RateLimiter {
	return &RateLimiter{
		limiter:  rate.NewLimiter(r, b),
		requests: requests,
	}
}

func (r *RateLimiter) CallAPI() {
	for i := 0; i < r.requests; i++ {
		err := r.limiter.Wait(context.Background())
		if err != nil {
			fmt.Println("Rate limit error:", err)
			return
		}

		// Make your API call here
		fmt.Println("API request:", i+1)
	}
}

// Define a type for email handler functions
type EmailHandlerFunc func(email *gmail.Message) error

// ProcessNewEmails retrieves new emails and processes them
func ProcessNewEmails(ctx context.Context, gmailService *gmail.Service, historyFile string, handlers []EmailHandlerFunc) error {
	// Read the last historyId from the file
	lastHistoryId, err := readHistoryId(historyFile)
	if err != nil {
		return fmt.Errorf("unable to read last historyId from file: %v", err)
	}

	histories, err := integration.GetHistoryList(gmailService, "me", lastHistoryId)
	if err != nil {
		return fmt.Errorf("unable to get histories %v", err)
	}
	// Create a WaitGroup to wait for all email handler functions to complete
	var wg sync.WaitGroup
	// Process each message returned by the history API
	for _, h := range histories.History {
		for _, m := range h.Messages {
			// Retrieve only the message headers to limit the size of the response
			msg, err := gmailService.Users.Messages.Get("me", m.Id).Format("full").Do()
			if err != nil {
				return fmt.Errorf("unable to retrieve message %v: %v", m.Id, err)
			}
			fmt.Println(msg.Snippet)

			// Process the email content with each handler function to determine if it meets the criteria
			for _, handler := range handlers {
				wg.Add(1)
				go func(h EmailHandlerFunc) {
					defer wg.Done()
					err := h(msg)
					if err != nil {
						fmt.Printf("error processing email %v: %v", msg.Id, err)
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
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Write the updated lastHistoryId to the file
	err = writeHistoryId(historyFile, lastHistoryId)
	if err != nil {
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
