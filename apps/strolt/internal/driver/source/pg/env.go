package pg

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func (i *PgDump) getEnv() []string {
	env := []string{}

	for envName, envValue := range i.env {
		env = append(env, fmt.Sprintf("%s=%s", envName, envValue))
	}

	if i.config.Username != "" {
		env = append(env, fmt.Sprintf("PGUSER=%s", i.config.Username))
	}

	if i.config.Password != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", i.config.Password))
	}

	if i.config.Port != 0 {
		env = append(env, fmt.Sprintf("PGPORT=%d", i.config.Port))
	}

	if i.config.Host != "" {
		env = append(env, fmt.Sprintf("PGHOST=%s", i.config.Host))
	}

	return env
}

func (i *PgDump) SetEnv(env interface{}) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return err
	}

	return nil
}
