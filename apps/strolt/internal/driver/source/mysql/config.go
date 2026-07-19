package mysql

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// TLS modes supported by the driver.
const (
	// TLSModeDisabled disables TLS entirely (`--skip-ssl`).
	TLSModeDisabled = "disabled"
	// TLSModeSkipVerify uses TLS but does not verify the server certificate
	// (`--skip-ssl-verify-server-cert`). Required for MySQL servers with
	// auto-generated self-signed certificates: the MariaDB client verifies
	// certificates by default since 11.4 and can do so automatically only
	// against MariaDB servers.
	TLSModeSkipVerify = "skip-verify"
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
	TLS      string `yaml:"tls"`
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
	switch i.config.TLS {
	case "", TLSModeDisabled, TLSModeSkipVerify:
		return nil
	default:
		return fmt.Errorf("unsupported tls mode %q, expected %q or %q", i.config.TLS, TLSModeDisabled, TLSModeSkipVerify)
	}
}

func (i *MySQL) getCommonArgs() []string {
	args := []string{}

	switch i.config.TLS {
	case TLSModeDisabled:
		args = append(args, "--skip-ssl")
	case TLSModeSkipVerify:
		args = append(args, "--skip-ssl-verify-server-cert")
	}

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

	args := make([]string, 0, len(commonArgs)+2)
	args = append(args, commonArgs...)

	// The dump is fed to the client on stdin (see Restore) rather than through
	// the client-side `source` builtin: in non-interactive batch mode the client
	// aborts on the first failing statement and returns a non-zero exit code,
	// while `-e "source ..."` swallows per-statement errors and can exit 0 on a
	// partial restore.
	args = append(args, "-D", i.config.Database)

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
