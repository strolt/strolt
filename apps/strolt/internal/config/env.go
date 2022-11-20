package config

import (
	"os"
	"strings"
)

type EnvConfig struct {
	Tags []string
}

func getEnvConfig() Config {
	c := EnvConfig{}
	tags := os.Getenv("STROLT_GLOBAL_TAGS")

	if tags != "" {
		c.Tags = strings.Split(strings.ReplaceAll(tags, " ", ""), ",")
	}

	return Config{
		Tags: c.Tags,
	}
}
