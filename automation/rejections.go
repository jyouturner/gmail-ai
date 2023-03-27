package automation

import (
	"fmt"
	"log"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"google.golang.org/api/gmail/v1"
)

func ProcessNewEmails(gmailService *gmail.Service, chatGPTClient *integration.ChatGPTClient) {
	// List unread emails
	results, err := gmailService.Users.Messages.List("me").Q("is:unread").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	for _, message := range results.Messages {
		msg, err := gmailService.Users.Messages.Get("me", message.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", message.Id, err)
		}
		fmt.Println(msg.Snippet)
		// Process the email content with your function to determine if it's a rejection
		// If it's a rejection, mark it as read, archive it, or apply a label
		isRejection, err := chatGPTClient.IsRejectionEmail(msg.Snippet)
		if err != nil {
			log.Fatalf("Error while checking if email is a rejection: %v", err)
		}

		if isRejection {
			// Mark the email as read, archive it, or apply a label
			//err := setLabel(gmailService, message.Id, "Rejections", true, false)
			//if err != nil {
			//	log.Fatalf("Error setting label on message %v: %v", message.Id, err)
			//}
			fmt.Println("Rejection email found!")
		}

	}
}
