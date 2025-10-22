package pg

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatCustom    Format = "c"
	FormatDirectory Format = "d"
	FormatTar       Format = "t"
	FormatPlainText Format = "p"
)

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

	BackupParams  string `yaml:"backup_params"`
	RestoreParams string `yaml:"restore_params"`
	CommonParams  string `yaml:"common_params"`
}

func (i *PgDump) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return err
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

	return nil
}

func (i *PgDump) getCommonArgs() []string {
	args := []string{}

	if i.config.CommonParams != "" {
		args = append(args, i.config.CommonParams)
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
		args = append(args, i.config.BackupParams)
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
		args = append(args, i.config.RestoreParams)
	}

	args = append(args, i.getCommonArgs()...)

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
