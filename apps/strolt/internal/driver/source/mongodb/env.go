package mongodb

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func (i *MongoDB) getEnv() []string {
	env := []string{}

	for envName, envValue := range i.env {
		env = append(env, fmt.Sprintf("%s=%q", envName, envValue))
	}

	return env
}

func (i *MongoDB) SetEnv(env interface{}) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &i.env)
}
