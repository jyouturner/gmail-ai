package automation

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/jyouturer/gmail-ai/internal/logging"
)

func TestMain(m *testing.M) {
	// Set up the logger for testing
	logger, err := logging.NewLogger()
	if err != nil {
		// Handle error
		panic(err)
	}
	logging.Logger = logger

	// Run the tests
	code := m.Run()

	// Clean up any resources here

	os.Exit(code)
}

func TestRejection(t *testing.T) {
	//packagetest.TestMain(nil)
	// Create a connection pool with 10 GRPCClient objects
	cp, err := integration.NewConnectionPool("localhost:50051", 2, time.Second*10)
	if err != nil {
		t.Errorf("Error creating connection pool: %v", err)
	}
	defer cp.Close()

	rc, err := cp.GetGRPCClient()

	defer cp.ReturnGRPCClient(rc)
	if err != nil {
		t.Errorf("Error get rejection chjeck: %v", err)
	}
	isRejection := IsRejection(context.Background(), rc, "email text goes here")

	fmt.Printf("res: %v\n", isRejection)
}
