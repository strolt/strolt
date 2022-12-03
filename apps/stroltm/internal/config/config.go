package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Instances []Instance `yaml:"instances"`
}

type Instance struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

var config Config

func Scan() error {
	data, err := os.ReadFile("./testdata/stroltm.yml")
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	return nil
}

func Get() Config {
	return config
}
