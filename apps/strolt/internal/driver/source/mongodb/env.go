package mongodb

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func (i *MongoDB) getEnv() []string {
	env := make([]string, 0, len(i.env))

	for envName, envValue := range i.env {
		env = append(env, fmt.Sprintf("%s=%q", envName, envValue))
	}

	return env
}

// SetEnv parses the driver environment variables.
func (i *MongoDB) SetEnv(env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}

	return nil
}
