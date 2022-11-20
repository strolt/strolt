package mysql

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

const FileNamePrefix = "strolt_driver_mysql"

type MySQL struct {
	logger *logger.Logger
	config Config
}

func New() *MySQL {
	return &MySQL{}
}

func (i *MySQL) getFileName() string {
	return fmt.Sprintf("%s.sql", FileNamePrefix)
}

func (i *MySQL) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *MySQL) IsEmpty() (bool, error) {
	return true, nil
}
