// logging/logging.go
package logging

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
