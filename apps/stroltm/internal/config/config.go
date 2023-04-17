package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	API     API     `yaml:"api"`
	Strolt  Strolt  `yaml:"strolt"`
	Stroltp Stroltp `yaml:"stroltp"`
}

type API struct {
	Users map[string]User `yaml:"users"`
}

type User struct {
	Password string `yaml:"password"`
}

type Strolt struct {
	Instances map[string]Instance `yaml:"instances"`
}

type Stroltp struct {
	Instances map[string]Instance `yaml:"instances"`
}

type Instance struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var config Config

func Load(pathname string) error {
	data, err := os.ReadFile(pathname)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &config)
}

func Get() Config {
	return config
}

func GetUsers() map[string]string {
	users := map[string]string{}

	for username, user := range Get().API.Users {
		users[username] = user.Password
	}

	return users
}
