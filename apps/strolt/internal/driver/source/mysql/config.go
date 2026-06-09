package mysql

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config describes the MySQL source driver configuration.
type Config struct {
	BinPathMySQL     string `yaml:"bin_path_mysql"`
	BinPathMySQLDump string `yaml:"bin_path_mysqldump"`

	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// SetConfig parses and validates the driver configuration.
func (i *MySQL) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return i.validateConfig()
}

func (i *MySQL) validateConfig() error {
	return nil
}

func (i *MySQL) getCommonArgs() []string {
	args := []string{}

	if i.config.Host != "" {
		args = append(args, "-h", i.config.Host)
	}

	if i.config.Port != 0 {
		args = append(args, "-P", strconv.Itoa(i.config.Port))
	}

	if i.config.Username != "" {
		args = append(args, "-u", i.config.Username)
	}

	if i.config.Password != "" {
		args = append(args, "-p"+i.config.Password)
	}

	return args
}

func (i *MySQL) getBackupArgs() []string {
	args := []string{}

	args = append(args, i.getCommonArgs()...)

	args = append(args, "--no-tablespaces")

	args = append(args, "--result-file="+i.getFileName())

	if i.config.Database != "" {
		args = append(args, i.config.Database)
	}

	return args
}

func (i *MySQL) getRestoreArgs() []string {
	commonArgs := i.getCommonArgs()

	args := make([]string, 0, len(commonArgs)+4)
	args = append(args, commonArgs...)

	args = append(args, "-D", i.config.Database)

	args = append(args, "-e", "source "+i.getFileName())

	return args
}

func (i *MySQL) getBinMySQLDump() string {
	if i.config.BinPathMySQLDump != "" {
		return i.config.BinPathMySQLDump
	}

	return "/usr/bin/mariadb-dump"
}

func (i *MySQL) getBinMySQL() string {
	if i.config.BinPathMySQL != "" {
		return i.config.BinPathMySQL
	}

	return "/usr/bin/mariadb"
}
