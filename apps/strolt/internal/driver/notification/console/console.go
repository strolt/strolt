package console

import (
	"encoding/json"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

type Console struct {
	logger *logger.Logger
	config interface{}
}

func New() *Console {
	return &Console{}
}

func (i *Console) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *Console) SetConfig(config interface{}) error {
	if err := validateConfig(config); err != nil {
		return err
	}

	i.config = config

	return nil
}

func validateConfig(config interface{}) error {
	return nil
}

func (i *Console) Send(ctx context.Context) {
	data, err := json.Marshal(ctx)
	if err == nil {
		i.logger.Info(string(data))
	}
}
