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

func Port() int {
	return resultConfig.Port
}

func Host() string {
	return resultConfig.Host
}

func IsDebug() bool {
	return resultConfig.LogLevel == logger.LogLevelDebug || resultConfig.LogLevel == logger.LogLevelTrace
}

func IsAPILogEnabled() bool {
	return resultConfig.IsAPILogEnabled
}
