package integration

import (
	"os"
	"testing"
)

func TestChatGPTAPI(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "CHATGPT_API_KEY")
	chatgpt := NewChatGPTClient(os.Getenv("CHATGPT_API_KEY"))
	rejected, err := chatgpt.IsRejectionEmail("Good")
	if err != nil {
		t.Errorf("error checking rejection email: %v", err)
	}
	if rejected {
		t.Errorf("email should not be rejected")
	}

}

func TestChatGPTAPIRejected(t *testing.T) {
	email := `Dear Applicant,

	Thank you again for your interest in pursuing career opportunities at GEICO. We appreciate the time you've spent getting to know our company.
	
	Your application and the information you've provided have been carefully considered. Unfortunately, you have not been selected for: Engineering Manager - Data (REMOTE) .
	
	You are welcome to re-apply for other career opportunities that you are qualified for six months from the date of this email.
	We wish you every success in your future endeavors.
	
	Sincerely,
	
	GEICO Hiring Team
	**Please do not reply to this message as it has been sent from an unmonitored account.** `
	ignoreTestWithoutEnvironmentVariables(t, "CHATGPT_API_KEY")
	chatgpt := NewChatGPTClient(os.Getenv("CHATGPT_API_KEY"))
	rejected, err := chatgpt.IsRejectionEmail(email)
	if err != nil {
		t.Errorf("error checking rejection email: %v", err)
	}
	if !rejected {
		t.Errorf("email should be rejected")
	}

}
