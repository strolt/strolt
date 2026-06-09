// Package console implements a notification driver that writes messages to the log.
package console

import (
	"encoding/json"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/shared/logger"
)

// Console is a notification driver that logs the operation context.
type Console struct {
	logger *logger.Logger
	config any
}

// New creates a Console driver.
func New() *Console {
	return &Console{}
}

// SetLogger sets the logger used by the driver.
func (i *Console) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// SetConfig validates and stores the driver configuration.
func (i *Console) SetConfig(config any) error {
	if err := validateConfig(config); err != nil {
		return err
	}

	i.config = config

	return nil
}

// Send logs the operation context as JSON.
func (i *Console) Send(ctx context.Context) {
	data, err := json.Marshal(ctx)
	if err == nil {
		i.logger.Info(string(data))
	}
}

func validateConfig(config any) error {
	return nil
}
