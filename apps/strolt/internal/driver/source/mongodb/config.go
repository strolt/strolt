package mongodb

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/util/shellargs"

	"gopkg.in/yaml.v3"
)

// ArchiveFileName is the name of the archive file produced by mongodump.
const ArchiveFileName = "./strolt_driver_mongodb.gz"

// Config describes the MongoDB source driver configuration.
type Config struct {
	BinPathMongoDump    string `yaml:"bin_path_mongodump"`
	BinPathMongoRestore string `yaml:"bin_path_mongorestore"`

	URI      string `yaml:"uri"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	// BackupParams, RestoreParams and CommonParams are shell-like extra CLI
	// options for the respective mongo tool: whitespace separates options and
	// single or double quotes group a value (so it may contain spaces).
	// Broken quoting is rejected at config load.
	BackupParams  string `yaml:"backup_params"`
	RestoreParams string `yaml:"restore_params"`
	CommonParams  string `yaml:"common_params"`
}

// SetConfig parses the driver configuration.
func (i *MongoDB) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if err := i.validateConfig(); err != nil {
		return err
	}

	return nil
}

func (i *MongoDB) validateConfig() error {
	// Params reach the mongo tool as separate argv elements; a broken quoting
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

func (i *MongoDB) getCommonArgs() []string {
	args := []string{}

	if i.config.CommonParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the mongo tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.CommonParams)...)
	}

	args = append(args, "--gzip")

	args = append(args, "--archive="+ArchiveFileName)

	if i.config.URI != "" {
		args = append(args, fmt.Sprintf("--uri=%q", i.config.URI))
	} else {
		if i.config.Host != "" {
			args = append(args, fmt.Sprintf("--host=%q", i.config.Host))
		}

		if i.config.Port != 0 {
			args = append(args, fmt.Sprintf("--port=%d", i.config.Port))
		}
	}

	if i.config.Database != "" {
		args = append(args, "--db="+i.config.Database)
	}

	if i.config.Username != "" {
		args = append(args, fmt.Sprintf("--username=%q", i.config.Username))
	}

	if i.config.Password != "" {
		args = append(args, fmt.Sprintf("--password=%q", i.config.Password))
	}

	return args
}

func (i *MongoDB) getBackupArgs() []string {
	args := []string{}

	if i.config.BackupParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the mongo tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.BackupParams)...)
	}

	args = append(args, i.getCommonArgs()...)

	return args
}

func (i *MongoDB) getRestoreArgs() []string {
	args := []string{}

	if i.config.RestoreParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes the mongo tool abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.RestoreParams)...)
	}

	args = append(args, i.getCommonArgs()...)

	return args
}

func (i *MongoDB) getBinMongoDump() string {
	if i.config.BinPathMongoDump != "" {
		return i.config.BinPathMongoDump
	}

	return "/usr/bin/mongodump"
}

func (i *MongoDB) getBinMongoRestore() string {
	if i.config.BinPathMongoRestore != "" {
		return i.config.BinPathMongoRestore
	}

	return "/usr/bin/mongorestore"
}
