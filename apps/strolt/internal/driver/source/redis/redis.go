// Package redis provides a Redis/Valkey source driver based on redis-cli --rdb.
package redis

import (
	"github.com/strolt/strolt/shared/logger"
)

// Redis is a source driver that backs up Redis and Valkey instances.
type Redis struct {
	logger *logger.Logger
	config Config
	env    map[string]string
}

// New creates a new Redis driver instance.
func New() *Redis {
	return &Redis{}
}

// SetLogger sets the logger used by the driver.
func (i *Redis) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// IsEmpty reports whether the source contains no data.
func (i *Redis) IsEmpty() (bool, error) {
	return true, nil
}
