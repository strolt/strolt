package mongodb

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const ArchiveFileName = "./strolt_driver_mongodb.gz"

type Config struct {
	BinPathMongoDump    string `yaml:"bin_path_mongodump"`
	BinPathMongoRestore string `yaml:"bin_path_mongorestore"`

	URI      string `yaml:"uri"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	BackupParams  string `yaml:"backup_params"`
	RestoreParams string `yaml:"restore_params"`
	CommonParams  string `yaml:"common_params"`
}

func (i *MongoDB) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &i.config)
}

func (i *MongoDB) getCommonArgs() []string {
	args := []string{}

	if i.config.CommonParams != "" {
		args = append(args, i.config.CommonParams)
	}

	args = append(args, "--gzip")

	args = append(args, fmt.Sprintf("--archive=%s", ArchiveFileName))

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
		args = append(args, fmt.Sprintf("--db=%s", i.config.Database))
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
		args = append(args, i.config.BackupParams)
	}

	args = append(args, i.getCommonArgs()...)

	return args
}

func (i *MongoDB) getRestoreArgs() []string {
	args := []string{}

	if i.config.RestoreParams != "" {
		args = append(args, i.config.RestoreParams)
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
