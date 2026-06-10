// Package config loads and exposes the stroltp YAML configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root of the stroltp configuration file.
type Config struct {
	API    API    `yaml:"api"`
	Strolt Strolt `yaml:"strolt"`
}

// API holds the API server configuration.
type API struct {
	Users map[string]User `yaml:"users"`
}

// User describes credentials of an API user.
type User struct {
	Password string `yaml:"password"`
}

// Strolt holds the configuration of managed strolt instances.
type Strolt struct {
	Instances map[string]Instance `yaml:"instances"`
}

// Instance describes connection settings of a single strolt instance.
type Instance struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var config Config

// Load reads and parses the configuration file at pathname.
func Load(pathname string) error {
	data, err := os.ReadFile(pathname) //nolint:gosec // path comes from user-provided configuration
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	return nil
}

// Get returns the loaded configuration.
func Get() Config {
	return config
}

// GetUsers returns API users as a username to password map.
func GetUsers() map[string]string {
	users := map[string]string{}

	for username, user := range Get().API.Users {
		users[username] = user.Password
	}

	return users
}
