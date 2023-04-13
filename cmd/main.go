package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jyouturer/gmail-ai/automation"
	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func init() {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}
	logging.Logger = logger // Set the global logger instance
}

func main() {

	defer logging.Logger.Sync()

	var configFilePath string
	app := &cli.App{
		Name:  "gmail-ai",
		Usage: "use ChatGPT to automate your gmails",
		Commands: []*cli.Command{
			{
				Name:  "poll",
				Usage: "poll new emails and process them",
				Action: func(cCtx *cli.Context) error {
					poll(configFilePath)
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
		logging.Logger.Fatal("Error loading app", zap.Error(err))
	}

}

// poll polls new emails and processes them
func poll(configFilePath string) {
	// Load the configuration file
	config, err := automation.LoadConfig(configFilePath)
	if err != nil {
		logging.Logger.Fatal("Error loading config file", zap.Error(err))
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}
	// create process to handle rejection email
	rc, closeFunc, err := automation.NewRejectionChecker(config.GRPCService.URL, 10, 10)
	if err != nil {
		logging.Logger.Fatal("Error creating Rejection Checker", zap.Error(err))
	}
	defer closeFunc()
	// crate the gmail handler
	hc := automation.NewRejectionEmail(gmailService, rc)

	handlers := []automation.EmailHandlerFunc{
		hc.Process,
	}

	// Process new emails
	for {
		automation.ProcessNewEmails(context.Background(), gmailService, "history.txt", handlers)
		time.Sleep(60 * time.Second)
	}
}
