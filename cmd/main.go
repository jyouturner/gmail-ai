package main

import (
	"log"
	"os"
	"time"

	"github.com/jyouturer/gmail-ai/automation"
	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/urfave/cli/v2"
)

func main() {
	var configFilePath string
	app := &cli.App{
		Name:  "gmail-ai",
		Usage: "use ChatGPT to automate your gmails",
		Commands: []*cli.Command{
			{
				Name:  "label-rejection",
				Usage: "label rejection emails",
				Action: func(cCtx *cli.Context) error {
					labelRejections(configFilePath)
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       "config.json",
				Usage:       "path to the config file",
				Destination: &configFilePath,
				Required:    true,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func labelRejections(configFilePath string) {
	// Load the configuration file
	config, err := automation.LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}

	// Create ChatGPT API client
	chatGptClient := integration.NewChatGPTClient(config.ChatGPT.URL, config.ChatGPT.APIKey, integration.WithTimeout(time.Duration(config.ChatGPT.Timeout)*time.Second))

	// Process new emails
	automation.ProcessNewEmails(gmailService, chatGptClient)
}
