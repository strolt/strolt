// Package env parses strolt environment variables.
package env

import (
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/strolt/strolt/shared/logger"
)

type config struct {
	Host                 string          `env:"STROLT_HOST"                        envDefault:"0.0.0.0"`
	Port                 int             `env:"STROLT_PORT"                        envDefault:"8080"`
	GlobalTags           globalTags      `env:"STROLT_GLOBAL_TAGS"`
	LogLevel             logger.LogLevel `env:"STROLT_LOG_LEVEL"`
	IsAPILogEnabled      bool            `env:"STROLT_API_LOG_ENABLED"`
	IsWatchFilesDisabled bool            `env:"STROLT_DISABLE_WATCH_FILES_CHANGED"`
	PathData             string          `env:"STROLT_PATH_DATA"`
}

type globalTags []string

func (t *globalTags) UnmarshalText(text []byte) error {
	str := string(text)
	if str != "" {
		*t = strings.Split(strings.ReplaceAll(str, " ", ""), ",")
		return nil
	}

	return nil
}

var resultConfig config

// Scan parses environment variables and configures the log level.
func Scan() {
	if err := env.Parse(&resultConfig); err != nil {
		logger.New().Fatal(err)
	}

	switch resultConfig.LogLevel {
	case logger.LogLevelDebug:
		logger.SetLogLevel(logger.LogLevelDebug)
	case logger.LogLevelTrace:
		logger.SetLogLevel(logger.LogLevelTrace)
	case logger.LogLevelInfo:
	default:
		logger.SetLogLevel(logger.LogLevelInfo)
	}
}

// Port returns the configured API port.
func Port() int {
	return resultConfig.Port
}

// Host returns the configured API host.
func Host() string {
	return resultConfig.Host
}

// GlobalTags returns tags applied to all tasks.
func GlobalTags() []string {
	return resultConfig.GlobalTags
}

// IsDebug reports whether debug or trace logging is enabled.
func IsDebug() bool {
	return resultConfig.LogLevel == logger.LogLevelDebug || resultConfig.LogLevel == logger.LogLevelTrace
}

// IsAPILogEnabled reports whether API request logging is enabled.
func IsAPILogEnabled() bool {
	return resultConfig.IsAPILogEnabled
}

// IsWatchFilesDisabled reports whether config file watching is disabled.
func IsWatchFilesDisabled() bool {
	return resultConfig.IsWatchFilesDisabled
}

// PathData returns the configured data directory path.
func PathData() string {
	return resultConfig.PathData
}
