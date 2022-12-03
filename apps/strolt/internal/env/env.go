package env

import (
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

type config struct {
	Host       string     `env:"STROLT_HOST" envDefault:"0.0.0.0"`
	Port       int        `env:"STROLT_PORT" envDefault:"8080"`
	GlobalTags globalTags `env:"STROLT_GLOBAL_TAGS"`
	IsDebug    bool       `env:"STROLT_DEBUG"`
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
		logger.New().Error(err)
		os.Exit(1)
	}

	if resultConfig.IsDebug {
		logger.SetLogLevel(logger.LogLevelDebug)
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
	return resultConfig.IsDebug
}
