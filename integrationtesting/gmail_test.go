//go:build integrationtesting
// +build integrationtesting

package integrationtesting

import (
	"fmt"
	"os"
	"testing"

	integration "github.com/jyouturer/gmail-ai/integrations"
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
	gmailService, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

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
	gmailService, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	projectID := "theautomaticmanager"
	topicName := "incoming-gmails"
	err = integration.CallWatch(gmailService, projectID, topicName)
	if err != nil {
		t.Errorf("error calling watch: %v", err)
	}
}

func TestGetHistoryMessages(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		t.Errorf("Unable to get user profile: %v", err)
	}
	lastHistoryId := profile.HistoryId
	messages, err := integration.GetHistorieMessages(gmailService, "me", lastHistoryId)
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

func TestGetHistoryList(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		t.Errorf("Unable to get user profile: %v", err)
	}
	lastHistoryId := profile.HistoryId
	lastHistoryId, history, err := integration.GetHistoryList(gmailService, "me", lastHistoryId)
	if err != nil {
		t.Errorf("error getting histories: %v", err)
	}
	fmt.Printf("history: %v\n", history)
}

func TestGmailGetMessageById(t *testing.T) {
	ignoreTestWithoutEnvironmentVariables(t, "GMAIL_CREDENTIALS", "GMAIL_TOKEN")
	gmailService, err := integration.CreateGmailService(os.Getenv("GMAIL_CREDENTIALS"), os.Getenv("GMAIL_TOKEN"))

	if err != nil {
		t.Errorf("error creating gmail service: %v", err)
	}
	fmt.Printf("gmailService: %v\n", gmailService)
	messageID := "187707239af927a0"
	msg, err := gmailService.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		t.Errorf("unable to retrieve message %v: %v\n", messageID, err)
	}
	fmt.Printf("msg snippet: %v\n", msg.Snippet)
	text, err := integration.GetMessageCriticalContents(msg)
	if err != nil {
		t.Errorf("unable to parse message %v: %v\n", messageID, err)
	}
	fmt.Printf("msg: %v\n", text)

}
