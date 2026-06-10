// Package pg provides a PostgreSQL source driver based on pg_dump, pg_restore and psql.
package pg

import (
	"github.com/strolt/strolt/shared/logger"
)

// FileNamePrefix is the prefix of dump files produced by the driver.
const FileNamePrefix = "strolt_driver_pg"

// PgDump is a source driver that backs up and restores PostgreSQL databases.
//
//nolint:revive // keep existing exported name for backward compatibility
type PgDump struct {
	logger *logger.Logger
	config PgDumpConfig
	env    map[string]string
}

// New creates a new PgDump driver instance.
func New() *PgDump {
	return &PgDump{}
}

// SetLogger sets the logger used by the driver.
func (i *PgDump) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

// IsEmpty reports whether the source contains no data.
func (i *PgDump) IsEmpty() (bool, error) {
	return true, nil
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
