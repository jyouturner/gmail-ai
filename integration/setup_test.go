package integration

import "github.com/jyouturer/gmail-ai/internal/logging"

func init() {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}
	logging.Logger = logger // Set the global logger instance
}
