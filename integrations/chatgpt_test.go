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

func TestChatGPTAPIRejected2(t *testing.T) {
	email := `Thank you for taking the time to apply to Atlassian. We recently reviewed your application for the Senior Engineering Manager, Core Data Platform role and made the really tough decision not to move forward with your application at this time. We’re sorry it's not better news. Unfortunately, we aren’t able to provide more specific feedback at this stage.
 
	Thank you again, we really appreciate you taking the time to consider us. `
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
