package automation

import (
	"context"
	"fmt"
	"log"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"github.com/jyouturer/gmail-ai/internal/nlp"
	"google.golang.org/api/gmail/v1"
)

// define a handler
type Handler struct {
	GRPCClientPool *integration.ConnectionPool
	GmailService   *gmail.Service
}

// NewHandler creates a new Handler
func NewHandler(cp *integration.ConnectionPool, gmailService *gmail.Service) *Handler {
	// Create the client to call gRPC of the rejection classifier
	return &Handler{
		GRPCClientPool: cp,
		GmailService:   gmailService,
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
	// use NLP to extract the top 3 sentences of the email body
	topSentencens, err := nlp.ExtractTopSentenseFrom(3, text)
	if err != nil {
		return fmt.Errorf("unable to exract top sentences from message %v: %v", msg.Id, err)
	}
	rc, err := h.GRPCClientPool.GetGRPCClient()
	if err != nil {
		return fmt.Errorf("failed to get rejection check from pool %v", err)
	}
	defer h.GRPCClientPool.ReturnGRPCClient(rc)

	isRejection := IsRejection(ctx, rc, topSentencens)

	// If the email is a rejection, apply the specified label
	if isRejection {
		err := integration.SetLabel(h.GmailService, msg.Id, "Rejection", true, false)
		if err != nil {
			return fmt.Errorf("error setting label on message %v: %v", msg.Id, err)
		}
		logging.Logger.Info("Rejection email found!")
	}
	return nil
}

func IsRejection(ctx context.Context, grpcClient *integration.GRPCClient, text string) bool {
	req := &integration.ClassifyRequest{
		EmailText: text,
	}

	resp, err := grpcClient.Client.ClassifyEmail(ctx, req)
	if err != nil {
		log.Printf("failed to classify email: %v", err)
		return false
	}

	return resp.IsRejection
}
