package env

import (
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/strolt/strolt/shared/logger"
)

type config struct {
	Host                  string          `env:"STROLTP_HOST"                         envDefault:"0.0.0.0"`
	Port                  int             `env:"STROLTP_PORT"                         envDefault:"8080"`
	GlobalTags            globalTags      `env:"STROLTP_GLOBAL_TAGS"`
	LogLevel              logger.LogLevel `env:"STROLTP_LOG_LEVEL"`
	IsAPILogEnabled       bool            `env:"STROLTP_API_LOG_ENABLED"`
	IsWatchConfigDisabled bool            `env:"STROLTP_DISABLE_WATCH_CONFIG_CHANGED"`
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

func GlobalTags() []string {
	return resultConfig.GlobalTags
}

func IsDebug() bool {
	return resultConfig.LogLevel == logger.LogLevelDebug || resultConfig.LogLevel == logger.LogLevelTrace
}

func IsAPILogEnabled() bool {
	return resultConfig.IsAPILogEnabled
}

func IsWatchConfigDisabled() bool {
	return resultConfig.IsWatchConfigDisabled
}
