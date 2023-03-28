package integration

import (
	"fmt"
	"os"
	"testing"
)

func ignoreTestWithoutEnvironmentVariables(t *testing.T, envVars ...string) {
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("environment variable %s not set", envVar)
		}
	}
}

func TestGmailGetMessages(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	res, err := gmailService.Users.Labels.List("me").Do()
	if err != nil {

		t.Errorf("error listing labels: %v", err)
	}
	for _, label := range res.Labels {
		fmt.Printf("label: %v\n", label.Name)
	}
}

func TestCallWatch(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	projectID := "theautomaticmanager"
	topicName := "incoming-gmails"
	err = CallWatch(gmailService, projectID, topicName)
	if err != nil {
		t.Errorf("error calling watch: %v", err)
	}
}

func TestGetHistories(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	messages, err := GetHistories(gmailService, "me", uint64(5991011))
	if err != nil {
		t.Errorf("error getting histories: %v", err)
	}
	for _, msg := range messages {
		// Process the message (e.g., print its subject)
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				fmt.Printf("Message ID: %s, Subject: %s\n", msg.Id, header.Value)
			}
		}
	}

}
