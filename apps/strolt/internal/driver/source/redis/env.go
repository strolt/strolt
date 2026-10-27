package redis

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// getEnv builds the redis-cli process environment. Pairs are plain KEY=VALUE:
// exec does not shell-parse them, so quoting the value (as other drivers do)
// would make redis-cli treat the quotes as part of the secret.
func (i *Redis) getEnv() []string {
	env := make([]string, 0, len(i.env)+1)

	for envName, envValue := range i.env {
		env = append(env, envName+"="+envValue)
	}

	// redis-cli reads the password from REDISCLI_AUTH natively; keeping it
	// out of argv hides it from `ps`. Appended last so it wins over a
	// colliding env entry (exec keeps the last occurrence).
	if i.config.Password != "" {
		env = append(env, "REDISCLI_AUTH="+i.config.Password)
	}

	return env
}

// SetEnv parses the driver environment variables.
func (i *Redis) SetEnv(env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}

	return nil
}
