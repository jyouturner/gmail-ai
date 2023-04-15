package automation

import (
	"testing"

	"github.com/jyouturer/gmail-ai/internal/logging"
)

func TestMain(m *testing.M) {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}
	logging.Logger = logger // Set the global logger instance
}
