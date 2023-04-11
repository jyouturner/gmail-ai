package packagetest

import (
	"os"
	"testing"

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
