package redis

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/util/shellargs"

	"gopkg.in/yaml.v3"
)

// ArchiveFileName is the RDB target passed to redis-cli --rdb in manual mode.
const ArchiveFileName = "./strolt_driver_redis.rdb"

// rdbMagic starts every RDB file; the two version digits after it vary by
// server version, so only the header is validated.
const rdbMagic = "REDIS"

// Config describes the Redis source driver configuration.
type Config struct {
	BinPathRedisCli string `yaml:"bin_path_redis_cli"`

	URI      string `yaml:"uri"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	// BackupParams and CommonParams are shell-like extra CLI options for
	// redis-cli: whitespace separates options and single or double quotes
	// group a value (so it may contain spaces). Broken quoting is rejected
	// at config load.
	BackupParams string `yaml:"backup_params"`
	CommonParams string `yaml:"common_params"`
}

// SetConfig parses and validates the driver configuration.
func (i *Redis) SetConfig(config any) error {
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

	i.normalizeConfig()

	return nil
}

func (i *Redis) validateConfig() error {
	if i.config.URI == "" && i.config.Host == "" {
		return errors.New("host or uri is empty")
	}

	// Params reach redis-cli as separate argv elements; a broken quoting
	// would only surface at backup time, so it fails at config load instead.
	for field, params := range map[string]string{
		"backup_params": i.config.BackupParams,
		"common_params": i.config.CommonParams,
	} {
		if _, err := shellargs.Split(params); err != nil {
			return fmt.Errorf("invalid %s (%q): %w; options are shell-like: space-separated, quotes group a value", field, params, err)
		}
	}

	return nil
}

// normalizeConfig moves URI userinfo off the command line entirely: any
// credential there is visible in `ps`, and redis-cli treats userinfo
// (even a colon-less one, which it reads as the AUTH password) as
// authoritative, ignoring REDISCLI_AUTH. The password travels via
// REDISCLI_AUTH, the username via --user; explicit config fields win.
func (i *Redis) normalizeConfig() {
	if i.config.URI == "" {
		return
	}

	u, err := url.Parse(i.config.URI)
	if err != nil || u.User == nil {
		return
	}

	username := u.User.Username()
	password, hasPassword := u.User.Password()

	if hasPassword {
		if i.config.Password == "" {
			i.config.Password = password
		}

		if i.config.Username == "" {
			i.config.Username = username
		}
	} else if username != "" {
		// Colon-less userinfo is the AUTH password in redis-cli semantics,
		// not a username; keep that meaning when moving it off the argv.
		if i.config.Password == "" {
			i.config.Password = username
		}
	}

	u.User = nil
	i.config.URI = u.String()
}

// getFileName is the artifact name exposed as the pipe filename (restic
// --stdin-filename); manual mode writes ArchiveFileName instead.
func (i *Redis) getFileName() string {
	return strings.TrimPrefix(ArchiveFileName, "./")
}

func (i *Redis) getCommonArgs() []string {
	args := []string{}

	if i.config.URI != "" {
		args = append(args, "-u", i.config.URI)
	} else {
		args = append(args, "-h", i.config.Host)

		// No default port arg: redis-cli itself defaults to 6379.
		if i.config.Port != 0 {
			args = append(args, "-p", strconv.Itoa(i.config.Port))
		}
	}

	if i.config.Username != "" {
		args = append(args, "--user", i.config.Username)
	}

	if i.config.CommonParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes redis-cli abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.CommonParams)...)
	}

	return args
}

// getBackupArgs builds the redis-cli argv for a backup targeting the given
// RDB destination: the archive path in manual mode, "-" for a pipe.
func (i *Redis) getBackupArgs(target string) []string {
	args := []string{}

	if i.config.BackupParams != "" {
		// Split into separate argv elements: passing "--opt-a --opt-b" as
		// one element makes redis-cli abort with an invalid-option error.
		args = append(args, shellargs.MustSplit(i.config.BackupParams)...)
	}

	args = append(args, i.getCommonArgs()...)
	args = append(args, "--rdb", target)

	return args
}

func (i *Redis) getBinRedisCli() string {
	if i.config.BinPathRedisCli != "" {
		return i.config.BinPathRedisCli
	}

	return "/usr/bin/redis-cli"
}
