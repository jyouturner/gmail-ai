package automation

import (
	"context"
	"fmt"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"google.golang.org/api/gmail/v1"
)

// define a handler
type Handler struct {
	RejectionCheckPool *integration.ConnectionPool
	GmailService       *gmail.Service
}

// NewHandler creates a new Handler
func NewHandler(cp *integration.ConnectionPool, gmailService *gmail.Service) *Handler {
	// Create the client to call gRPC of the rejection classifier
	return &Handler{
		RejectionCheckPool: cp,
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
	rc, err := h.RejectionCheckPool.GetRejectionCheck()
	if err != nil {
		return fmt.Errorf("failed to get rejection check from pool %v", err)
	}
	defer h.RejectionCheckPool.ReturnRejectionCheck(rc)

	isRejection := rc.IsRejection(ctx, text)

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
