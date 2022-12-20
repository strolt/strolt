package env

import (
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/strolt/strolt/shared/logger"
)

type config struct {
	Host    string `env:"STROLTM_HOST" envDefault:"0.0.0.0"`
	Port    int    `env:"STROLTM_PORT" envDefault:"8080"`
	IsDebug bool   `env:"STROLTM_DEBUG"`
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

func IsDebug() bool {
	return resultConfig.IsDebug
}
