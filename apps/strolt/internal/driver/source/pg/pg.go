package pg

import (
	"github.com/strolt/strolt/shared/logger"
)

const FileNamePrefix = "strolt_driver_pg"

//nolint:revive
type PgDump struct {
	logger *logger.Logger
	config PgDumpConfig
	env    map[string]string
}

func New() *PgDump {
	return &PgDump{}
}

func (i *PgDump) getFileName() string {
	filename := FileNamePrefix
	if i.config.Format == "d" {
		return filename + "_directory"
	}

	if i.config.Format == "c" {
		return filename + ".custom"
	}

	if i.config.Format == "t" {
		return filename + ".tar"
	}

	return filename + ".sql"
}

func (i *PgDump) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *PgDump) IsEmpty() (bool, error) {
	return true, nil
}
