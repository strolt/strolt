package mongodb

import (
	"github.com/strolt/strolt/shared/logger"
)

type MongoDB struct {
	logger *logger.Logger
	config Config
	env    map[string]string
}

func New() *MongoDB {
	return &MongoDB{}
}

func (i *MongoDB) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *MongoDB) IsEmpty() (bool, error) {
	return true, nil
}
