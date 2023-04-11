package automation

import (
	"context"
	"fmt"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"google.golang.org/api/gmail/v1"
)

// define a handler
type Handler struct {
	RejectionClassfier *integration.RejectionCheck
	GmailService       *gmail.Service
}

// NewHandler creates a new Handler
func NewHandler(rejectionCheck *integration.RejectionCheck, gmailService *gmail.Service) *Handler {
	return &Handler{
		RejectionClassfier: rejectionCheck,
		GmailService:       gmailService,
	}
}

// Implement the HandleRejection method of the EmailHandlerFunc interface
func (h *Handler) HandleRejection(ctx context.Context, msg *gmail.Message) error {
	// Use ChatGPT to determine if the email is a rejection
	text, err := integration.GetMessage(msg)
	if err != nil {
		return fmt.Errorf("unable to parse message %v: %v", msg.Id, err)
	}
	if text == "" {
		return fmt.Errorf("empty message %v", msg.Id)
	}
	isRejection := h.RejectionClassfier.IsRejection(ctx, text)

	// If the email is a rejection, apply the specified label
	if isRejection {
		err := integration.SetLabel(h.GmailService, msg.Id, "Rejection", true, false)
		if err != nil {
			return fmt.Errorf("error setting label on message %v: %v", msg.Id, err)
		}
		fmt.Println("Rejection email found!")
	}
	return nil
}
