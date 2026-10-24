package pg

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/util/shellargs"

	"gopkg.in/yaml.v3"
)

// Format is a pg_dump output format identifier.
type Format string

// Supported pg_dump output formats.
const (
	FormatCustom    Format = "c"
	FormatDirectory Format = "d"
	FormatTar       Format = "t"
	FormatPlainText Format = "p"
)

// PgDumpConfig describes the PostgreSQL source driver configuration.
//
//nolint:revive // keep existing exported name for backward compatibility
type PgDumpConfig struct {
	BinPathPgDump    string `yaml:"bin_path_pg_dump"`
	BinPathPgRestore string `yaml:"bin_path_pg_restore"`
	BinPathPsql      string `yaml:"bin_path_psql"`

	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Format   Format `yaml:"format"` // -F, --format=c|d|t|p         output file format (custom, directory, tar, plain text (default))

	// BackupParams, RestoreParams and CommonParams are shell-like extra CLI
	// options for the respective pg tool: whitespace separates options and
	// single or double quotes group a value (so it may contain spaces).
	// Broken quoting is rejected at config load.
	BackupParams  string `yaml:"backup_params"`
	RestoreParams string `yaml:"restore_params"`
	CommonParams  string `yaml:"common_params"`
}

// SetConfig parses and validates the driver configuration.
func (i *PgDump) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return i.validateConfig()
}

func (i *PgDump) validateConfig() error {
	if i.config.Format != "" {
		if i.config.Format != FormatCustom &&
			i.config.Format != FormatTar &&
			i.config.Format != FormatDirectory &&
			i.config.Format != FormatPlainText {
			return errors.New("not available format. available [c|t|p|d]")
		}
	}

	// Params reach the pg tool as separate argv elements; a broken quoting
	// would only surface at restore time, so it fails at config load instead.
	for field, params := range map[string]string{
		"backup_params":  i.config.BackupParams,
		"restore_params": i.config.RestoreParams,
		"common_params":  i.config.CommonParams,
	} {
		if _, err := shellargs.Split(params); err != nil {
			return fmt.Errorf("invalid %s (%q): %w; options are shell-like: space-separated, quotes group a value", field, params, err)
		}
	}

	return nil
}

func (i *PgDump) getCommonArgs() []string {
	args := []string{}

	if i.config.CommonParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the pg tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.CommonParams)...)
	}

	args = append(args, "--no-password")

	if i.config.Database != "" {
		args = append(args, "--dbname="+i.config.Database)
	}

	// if i.config.Host != "" {
	// 	args = append(args, fmt.Sprintf("--host=%s", i.config.Host))
	// }

	// if i.config.Port != 0 {
	// 	args = append(args, fmt.Sprintf("--port=%d", i.config.Port))
	// }

	// if i.config.Username != "" {
	// 	args = append(args, fmt.Sprintf("--username=%s", i.config.Username))
	// }

	return args
}

func (i *PgDump) getBackupArgs() []string {
	args := []string{}

	if i.config.BackupParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the pg tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.BackupParams)...)
	}

	args = append(args, i.getCommonArgs()...)

	if i.config.Format != "" {
		args = append(args, fmt.Sprintf("--format=%s", i.config.Format))
	}

	return args
}

func (i *PgDump) getRestoreArgs() []string {
	args := []string{}

	if i.config.RestoreParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the pg tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.RestoreParams)...)
	}

	args = append(args, i.getCommonArgs()...)

	// Abort on the first error so a partial restore (truncated dump, restoring
	// into a non-empty/mismatched database) fails with a non-zero exit code
	// instead of silently exiting 0 and being reported as successful.
	args = append(args, "--exit-on-error")

	return args
}

func (i *PgDump) getBinPgDump() string {
	if i.config.BinPathPgDump != "" {
		return i.config.BinPathPgDump
	}

	return "/usr/bin/pg_dump"
}

func (i *PgDump) getBinPgRestore() string {
	if i.config.BinPathPgRestore != "" {
		return i.config.BinPathPgRestore
	}

	return "/usr/bin/pg_restore"
}

func (i *PgDump) getBinPSQL() string {
	if i.config.BinPathPsql != "" {
		return i.config.BinPathPsql
	}

	return "/usr/bin/psql"
}
