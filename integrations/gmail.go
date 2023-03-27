package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func CreateGmailService(credsPath string, tokenPath string) (*gmail.Service, error) {
	// Read OAuth 2.0 credentials from the JSON file
	credsBytes, err := ioutil.ReadFile(credsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	// Create OAuth 2.0 config
	config, err := google.ConfigFromJSON(credsBytes, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Load or request a new token
	token, err := GetTokenFromJSON(tokenPath, config)
	if err != nil {
		token = RequestNewToken(config)
		SaveTokenToJSON(tokenPath, token)
	}
	//mail API client
	client := config.Client(context.Background(), token)

	// Create the Gmail service
	gmailService, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	return gmailService, nil
}

// SetLabel sets the given label on the given message, and optionally marks it as read and archives it
func SetLabel(gmailService *gmail.Service, messageID, labelName string, markAsRead bool, archive bool) error {
	// Get a list of available labels
	labels, err := gmailService.Users.Labels.List("me").Do()
	if err != nil {
		return err
	}

	// Find the label ID that corresponds to the given labelName
	var labelID string
	for _, label := range labels.Labels {
		if label.Name == labelName {
			labelID = label.Id
			break
		}
	}

	// Create the label if it doesn't exist
	if labelID == "" {
		label := &gmail.Label{
			Name:                  labelName,
			LabelListVisibility:   "labelShow",
			MessageListVisibility: "show",
		}
		newLabel, err := gmailService.Users.Labels.Create("me", label).Do()
		if err != nil {
			return err
		}
		labelID = newLabel.Id
	}

	// Set the label, mark as read, and archive the message
	modifyRequest := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}

	if markAsRead {
		modifyRequest.RemoveLabelIds = append(modifyRequest.RemoveLabelIds, "UNREAD")
	}
	if archive {
		modifyRequest.RemoveLabelIds = append(modifyRequest.RemoveLabelIds, "INBOX")
	}

	_, err = gmailService.Users.Messages.Modify("me", messageID, modifyRequest).Do()
	return err
}

func CallWatch(gmailService *gmail.Service, projectID string, topicName string) error {
	// Replace these with your own values
	userID := "me" // Use "me" to represent the authenticated user

	// Set up Gmail watch
	watchRequest := &gmail.WatchRequest{
		LabelIds:  []string{"INBOX"},
		TopicName: fmt.Sprintf("projects/%s/topics/%s", projectID, topicName),
	}

	watchResponse, err := gmailService.Users.Watch(userID, watchRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to set up Gmail watch: %v", err)
	}

	log.Printf("Gmail watch response: %+v\n", watchResponse)
	return nil
}

func GetHistories(gmailService *gmail.Service, userID string, startHistoryID uint64) ([]*gmail.Message, error) {
	messages := []*gmail.Message{}
	// Get Gmail history
	historyListCall := gmailService.Users.History.List(userID)
	historyListCall.StartHistoryId(startHistoryID)
	historyList, err := historyListCall.Do()
	if err != nil {
		log.Fatalf("Failed to get Gmail history: %v", err)
	}

	// Iterate through the history records
	for _, historyRecord := range historyList.History {
		// Fetch messages from the history record
		for _, message := range historyRecord.Messages {
			msg, err := gmailService.Users.Messages.Get(userID, message.Id).Do()
			if err != nil {
				log.Printf("Failed to get message with ID %s: %v", message.Id, err)
				continue
			}
			messages = append(messages, msg)
		}
	}
	return messages, nil
}
