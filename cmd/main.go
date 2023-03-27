package main

import (
	"log"

	"github.com/jyouturer/gmail-ai/automation"
	integration "github.com/jyouturer/gmail-ai/integrations"
)

func main() {
	// Load the configuration file
	config, err := automation.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}

	// Create ChatGPT API client
	chatGptClient := integration.NewChatGPTClient(config.ChatGPT.APIKey)

	// Process new emails
	automation.ProcessNewEmails(gmailService, chatGptClient)
}
