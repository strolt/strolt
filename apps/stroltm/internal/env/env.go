// Package env parses the Strolt Manager configuration from environment variables.
package env

import (
	"github.com/caarlos0/env/v6"
	"github.com/strolt/strolt/shared/logger"
)

type config struct {
	Host            string          `env:"STROLTM_HOST"            envDefault:"0.0.0.0"`
	Port            int             `env:"STROLTM_PORT"            envDefault:"8080"`
	LogLevel        logger.LogLevel `env:"STROLTM_LOG_LEVEL"`
	IsAPILogEnabled bool            `env:"STROLTM_API_LOG_ENABLED"`
}

var resultConfig config

// Scan parses environment variables and applies the log level.
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

// Port returns the API server port.
func Port() int {
	return resultConfig.Port
}

// Host returns the API server host.
func Host() string {
	return resultConfig.Host
}

// IsDebug reports whether the debug or trace log level is enabled.
func IsDebug() bool {
	return resultConfig.LogLevel == logger.LogLevelDebug || resultConfig.LogLevel == logger.LogLevelTrace
}

// IsAPILogEnabled reports whether API request logging is enabled.
func IsAPILogEnabled() bool {
	return resultConfig.IsAPILogEnabled
}
