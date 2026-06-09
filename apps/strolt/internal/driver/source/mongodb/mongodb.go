// Package mongodb provides a MongoDB source driver based on mongodump and mongorestore.
package mongodb

import (
	"github.com/strolt/strolt/shared/logger"
)

// MongoDB is a source driver that backs up and restores MongoDB databases.
type MongoDB struct {
	logger *logger.Logger
	config Config
	env    map[string]string
}

// New creates a new MongoDB driver instance.
func New() *MongoDB {
	return &MongoDB{}
}

// SetLogger sets the logger used by the driver.
func (i *MongoDB) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// IsEmpty reports whether the source contains no data.
func (i *MongoDB) IsEmpty() (bool, error) {
	return true, nil
}
