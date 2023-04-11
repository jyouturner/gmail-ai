package main

import (
	"context"
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
					for {
						labelRejections(context.Background(), configFilePath)
						time.Sleep(10 * time.Second)
					}

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

func labelRejections(ctx context.Context, configFilePath string) {
	// Load the configuration file
	config, err := automation.LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}

	// Create the client to call gRPC of the rejection classifier
	rejectionCheck := integration.NewRejectionCheck(config.RejectionCheck.URL)
	// crate the gmail handler
	handler := automation.NewHandler(rejectionCheck, gmailService)
	// Process new emails
	// Create a context with a timeout of 10 seconds
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	automation.ProcessNewEmails(ctxTimeout, gmailService, "history.txt", []automation.EmailHandlerFunc{handler.HandleRejection})
}
