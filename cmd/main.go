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
				Name:  "label-rejection",
				Usage: "label rejection emails",
				Action: func(cCtx *cli.Context) error {
					labelRejections(context.Background(), configFilePath)
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

// labelRejections labels rejection emails
func labelRejections(ctx context.Context, configFilePath string) {
	// Load the configuration file
	config, err := automation.LoadConfig(configFilePath)
	if err != nil {
		logging.Logger.Fatal("Error loading config file", zap.Error(err))
	}

	gmailService, err := integration.CreateGmailService(config.Gmail.Credentials, config.Gmail.Token)
	if err != nil {
		log.Fatalf("Error creating Gmail service: %v", err)
	}

	// Create a connection pool with 10 RejectionCheck objects
	cp, err := integration.NewConnectionPool(config.RejectionCheck.URL, 10, time.Second*10)
	if err != nil {
		log.Fatalf("Error creating connection pool: %v", err)
	}
	defer cp.Close()
	// crate the gmail handler
	handler := automation.NewHandler(cp, gmailService)

	// Process new emails
	// Create a context with a timeout of 10 seconds
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		automation.ProcessNewEmails(ctxTimeout, gmailService, "history.txt", []automation.EmailHandlerFunc{handler.HandleRejection})
		time.Sleep(10 * time.Second)
	}
}
