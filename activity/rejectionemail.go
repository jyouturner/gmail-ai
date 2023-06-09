package activity

import (
	"context"
	"fmt"

	"github.com/jyouturer/gmail-ai/datamodel"
	"github.com/jyouturer/gmail-ai/integration"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"github.com/jyouturer/gmail-ai/internal/nlp"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
)

// Check whether a email is rejection, then label it
type RejectionEmail struct {
	GmailService      *gmail.Service
	RejectionChecking RejectionChecking
}

// NewRejectionEmail creates a new RejectionEmail, given gmail client and rejection checker implementation
func NewRejectionEmail(gmailService *gmail.Service, rc RejectionChecking) *RejectionEmail {
	return &RejectionEmail{
		GmailService:      gmailService,
		RejectionChecking: rc,
	}
}

// RejectionChecking interface, it will be implemented by the RejectionChecker
type RejectionChecking interface {
	IsRejection(ctx context.Context, text string) (bool, error)
}

// HandleRejectionEmail implement the EmailRejectionEmailFunc interface, it will be called by Email Processor
func (h *RejectionEmail) Process(ctx context.Context, msg datamodel.Message) error {
	// Get the text of the email
	text := msg.Body

	// use NLP to extract the top 3 sentences of the email body
	topSentencens, err := nlp.ExtractTopSentenseFrom(3, text)
	if err != nil {
		return fmt.Errorf("unable to exract top sentences from message %v: %v", msg.ID, err)
	}

	res, err := h.RejectionChecking.IsRejection(ctx, topSentencens)
	if err != nil {
		return fmt.Errorf("error calling IsRejection gRPC: %v", err)
	}
	logging.Logger.Debug("IsRejection", zap.Bool("res", res))
	// If the email is a rejection, apply the specified label
	if res {
		err := integration.SetLabel(h.GmailService, msg.ID, "Rejection", true, false)
		if err != nil {
			return fmt.Errorf("error setting label on message %v: %v", msg.ID, err)
		}
	}
	return nil
}
