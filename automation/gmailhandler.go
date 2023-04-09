package automation

import (
	"fmt"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"google.golang.org/api/gmail/v1"
)

// define a handler
type Handler struct {
	ChatGPTClient *integration.ChatGPTClient
	GmailService  *gmail.Service
}

// NewHandler creates a new Handler
func NewHandler(chatGPTClient *integration.ChatGPTClient, gmailService *gmail.Service) *Handler {
	return &Handler{
		ChatGPTClient: chatGPTClient,
		GmailService:  gmailService,
	}
}

// Implement the HandleRejection method of the EmailHandlerFunc interface
func (h *Handler) HandleRejection(msg *gmail.Message, ch chan<- error) {
	// Use ChatGPT to determine if the email is a rejection
	isRejection, err := h.ChatGPTClient.IsRejectionEmail(msg.Snippet)
	if err != nil {
		ch <- fmt.Errorf("error while checking if email is a rejection: %v", err)
		return
	}

	// If the email is a rejection, apply the specified label
	if isRejection {
		err := integration.SetLabel(h.GmailService, msg.Id, "Rejection", true, false)
		if err != nil {
			ch <- fmt.Errorf("error setting label on message %v: %v", msg.Id, err)
			return
		}
		fmt.Println("Rejection email found!")
	}

	// If there were no errors, send a nil error to the channel to indicate success
	ch <- nil
}
