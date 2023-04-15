package main

import (
	"context"
	"log"
	"os"
	"time"

	activity "github.com/jyouturer/gmail-ai/activity"
	config "github.com/jyouturer/gmail-ai/config"
	integration "github.com/jyouturer/gmail-ai/integration"
	"github.com/jyouturer/gmail-ai/internal/logging"
	gmailservice "github.com/jyouturer/gmail-ai/messagesource"
	automation "github.com/jyouturer/gmail-ai/polling"
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
	config, err := config.NewConfigFromFile(configFilePath)
	if err != nil {
		logging.Logger.Fatal("Error loading config file", zap.Error(err))
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}
	// create process to handle rejection email
	rc, closeFunc, err := activity.NewRejectionChecker(config.GRPCService.URL, 10, 10)
	if err != nil {
		logging.Logger.Fatal("Error creating Rejection Checker", zap.Error(err))
	}
	defer closeFunc()
	// crate the gmail handler
	hc := activity.NewRejectionEmail(gmailService, rc)

	handlers := []automation.MessageHandlerFunc{
		hc.Process,
	}
	provider := automation.NewMessageProvider(gmailservice.NewGmailService(gmailService))
	history := automation.NewFileHistory("history.txt")
	// Process new emails
	for {
		provider.PollAndProcess(context.Background(), history, handlers)
		time.Sleep(60 * time.Second)
	}
}
