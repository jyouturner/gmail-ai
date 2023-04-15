package integrationtesting

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jyouturer/gmail-ai/activity"
	"github.com/jyouturer/gmail-ai/integration"
	"github.com/jyouturer/gmail-ai/messagesource"
)

func TestRejection(t *testing.T) {
	//packagetest.TestMain(nil)
	// Create a connection pool with 10 GRPCClient objects
	cp, err := integration.NewConnectionPool("localhost:50051", 2, time.Second*10)
	if err != nil {
		t.Errorf("Error creating connection pool: %v", err)
	}
	defer cp.Close()

	c, err := cp.GetGRPCClient()

	defer cp.ReturnGRPCClient(c)
	if err != nil {
		t.Errorf("Error get rejection chjeck: %v", err)
	}
	rc := activity.RejectionChecker{
		GRPCClientPool: cp,
	}
	isRejection, err := rc.IsRejection(context.Background(), "email text goes here")
	if err != nil {
		t.Errorf("Error calling IsRejection gRPC: %v", err)
	}
	fmt.Printf("res: %v\n", isRejection)
	res, err := rc.IsRejection(context.Background(), "Thank you for your interest in the Senior Engineering Manager, Quip role at Salesforce. Unfortunately, we are no longer hiring for this Your time and effort are greatly appreciated.")
	fmt.Printf("res: %v, err: %v", res, err)
}

func TestHandleRejection(t *testing.T) {
	//packagetest.TestMain(nil)
	// Create a connection pool with 10 GRPCClient objects
	cp, err := integration.NewConnectionPool("localhost:50051", 2, time.Second*10)
	if err != nil {
		t.Errorf("Error creating connection pool: %v", err)
	}
	defer cp.Close()

	rc := activity.RejectionChecker{
		GRPCClientPool: cp,
	}

	// create gmail service
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmail, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail client: %v", err)
	}
	gmailService := messagesource.NewGmailService(gmail)
	fmt.Printf("gmailService: %v\n", gmailService)
	messageID := "1877976684734b16"
	msg, err := gmailService.GetMessage("me", messageID)
	if err != nil {
		t.Errorf("unable to retrieve message %v: %v\n", messageID, err)
	}

	// Create the client to call gRPC of the rejection classifier
	h := activity.NewRejectionEmail(gmailService.Gmail, &rc)

	// Implement the HandleRejection method of the EmailHandlerFunc interface
	err = h.Process(context.Background(), msg)
	if err != nil {
		t.Errorf("Error handling rejection: %v", err)
	}
}
