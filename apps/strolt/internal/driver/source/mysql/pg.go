// Package mysql provides a MySQL/MariaDB source driver based on mysqldump and mysql client tools.
package mysql

import (
	"github.com/strolt/strolt/shared/logger"
)

// FileNamePrefix is the prefix of dump files produced by the driver.
const FileNamePrefix = "strolt_driver_mysql"

// MySQL is a source driver that backs up and restores MySQL/MariaDB databases.
type MySQL struct {
	logger *logger.Logger
	config Config
}

// New creates a new MySQL driver instance.
func New() *MySQL {
	return &MySQL{}
}

// SetLogger sets the logger used by the driver.
func (i *MySQL) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// IsEmpty reports whether the source contains no data.
func (i *MySQL) IsEmpty() (bool, error) {
	return true, nil
}

func (i *MySQL) getFileName() string {
	return FileNamePrefix + ".sql"
}
