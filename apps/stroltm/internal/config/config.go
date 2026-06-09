// Package config loads and exposes the Strolt Manager configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration of the Strolt Manager.
type Config struct {
	API     API     `yaml:"api"`
	Strolt  Strolt  `yaml:"strolt"`
	Stroltp Stroltp `yaml:"stroltp"`
}

// API holds the API server configuration.
type API struct {
	Users map[string]User `yaml:"users"`
}

// User holds the credentials of an API user.
type User struct {
	Password string `yaml:"password"`
}

// Strolt holds the configuration of Strolt instances.
type Strolt struct {
	Instances map[string]Instance `yaml:"instances"`
}

// Stroltp holds the configuration of Stroltp instances.
type Stroltp struct {
	Instances map[string]Instance `yaml:"instances"`
}

// Instance describes the connection settings of a managed instance.
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
