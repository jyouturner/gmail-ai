package integrationtesting

import (
	"os"
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

func ignoreTestWithoutEnvironmentVariables(t *testing.T, envVars ...string) {
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("environment variable %s not set", envVar)
		}
	}
}
