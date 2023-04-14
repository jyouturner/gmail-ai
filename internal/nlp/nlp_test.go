package nlp

import (
	"os"
	"testing"

	"github.com/jyouturer/gmail-ai/internal/logging"
)

func TestMain(m *testing.M) {
	// Do stuff BEFORE the tests!
	logging.Logger, _ = logging.NewLogger()
	// Run the tests
	exitVal := m.Run()
	// Do stuff AFTER the tests!
	os.Exit(exitVal)
}

func TestNLP(t *testing.T) {

	emailBody := `
	Thank you for applying for the Software Engineering Manager - Backend/API (Mobile team) opportunity. We appreciate your interest in becoming a part of the exciting things we’re up to here at Northwestern Mutual.

Although your background is impressive, we have decided to move forward with other candidates who more closely align with what is needed for the position.

We appreciate your interest in Northwestern Mutual, and we encourage you to continue to check out our careers page as new opportunities are posted often.

Thank you again and best of luck in your career search!

Northwestern Mutual, Talent Acquisition 
`

	res, err := ExtractTopSentenseFrom(3, emailBody)
	if err != nil {
		t.Errorf("Error calling IsRejection: %v", err)
	}
	t.Logf("res: %v", res)
}
