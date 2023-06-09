package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// CreateGmailService creates a new Gmail service
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
	} else if token.Expiry.Before(time.Now()) {
		tokenSource := config.TokenSource(context.Background(), token)
		token, err = tokenSource.Token()
		if err != nil {
			log.Fatalf("Unable to refresh token: %v", err)
		}
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

// CallWatch sets up a Gmail watch for the given topic
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

// GetHistoriesList returns a list of history records that have changed since the last historyId
func GetHistoryList(gmailService *gmail.Service, userID string, lastHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error) {
	// If lastHistoryId is 0, perform a full sync
	var history *gmail.ListHistoryResponse
	if lastHistoryId == 0 {
		logging.Logger.Info("getting history for the first time from profie")
		profile, err := gmailService.Users.GetProfile(userID).Do()
		if err != nil {
			log.Fatalf("Unable to get user profile: %v", err)
		}
		lastHistoryId = profile.HistoryId
	}
	// List history records for messages that have changed since the last historyId
	logging.Logger.Info("last history id", zap.Uint64("last history id", lastHistoryId))
	history, err := gmailService.Users.History.List(userID).StartHistoryId(lastHistoryId).Do()
	if err != nil {
		return lastHistoryId, history, fmt.Errorf("unable to retrieve history list: %v", err)
	}
	return lastHistoryId, history, nil
}

// GetHistorieMessages returns a list of messages that have changed since the last historyId
func GetHistorieMessages(gmailService *gmail.Service, userID string, startHistoryID uint64) ([]*gmail.Message, error) {
	messages := []*gmail.Message{}
	// Get Gmail history
	_, historyList, err := GetHistoryList(gmailService, userID, startHistoryID)
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

// GetMessageCriticalContents returns the text content of the given message
func GetMessageCriticalContents(msg *gmail.Message) (string, error) {
	logging.Logger.Debug("GetMessage", zap.String("message", fmt.Sprintf("%+v", msg)))
	// Parse the message payload to get the text content
	payload := msg.Payload
	if payload != nil {
		// Check if the message is a multipart message
		if len(payload.Parts) > 0 {
			logging.Logger.Debug("Multipart Message", zap.Int("Parts", len(payload.Parts)))
			for i, part := range payload.Parts {

				// Check if the part is a text/plain or text/html part
				logging.Logger.Debug("Multipart Message", zap.Int("Part", i), zap.String("mimetype", part.MimeType))
				if part.MimeType == "text/plain" || part.MimeType == "text/html" {
					// Decode the part body to get the text content

					partBytes, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err != nil {
						return "", err
					}
					text := string(partBytes)
					logging.Logger.Debug("Multipart Message is plain text or html", zap.String("text", text))
					// Do something with the message text
					return text, nil
				} else if part.MimeType == "multipart/related" {
					// Loop through each part of the multipart/related body
					for _, subPart := range part.Parts {
						// Loop through each header in the part
						var contentType, contentDisposition string
						for _, header := range subPart.Headers {
							// Check if the header is the Content-Type header
							logging.Logger.Debug("header", zap.String("name", header.Name), zap.String("value", header.Value))
							if header.Name == "Content-Type" {
								// Get the content type of the part
								contentType = header.Value
							}

							// Check if the header is the Content-Disposition header
							if header.Name == "Content-Disposition" {
								// Get the content disposition of the part
								contentDisposition = header.Value
							}
						}

						// Check if the part contains the data you're looking for
						if strings.HasPrefix(contentType, "image/") && strings.Contains(contentDisposition, "attachment") {
							// Extract the data from the part
							data, err := base64.URLEncoding.DecodeString(subPart.Body.Data)
							if err != nil {
								return "", err
							}
							return string(data), nil
						}
					}

				} else {
					logging.Logger.Info("Multipart Message is not plain text", zap.String("mimetype", part.MimeType))

				}
			}
		} else {
			// If the message is not a multipart message, it may be a plain text message
			// with no MIME type specified
			if payload.MimeType == "text/plain" || payload.MimeType == "text/html" {
				// Decode the message body to get the text content
				bodyBytes, err := base64.URLEncoding.DecodeString(payload.Body.Data)
				if err != nil {
					// Handle error
					return "", err
				}
				text := string(bodyBytes)
				// Do something with the message text
				return text, nil
			} else {
				logging.Logger.Info("Message is not plain text", zap.String("mimetype", payload.MimeType))
				return "", nil
			}
		}
	}
	return "", nil
}
