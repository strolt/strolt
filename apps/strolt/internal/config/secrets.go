package config

import (
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"
)

func (c *Config) replaceSecrets() error {
	if err := c.replaceSecretsNotifications(); err != nil {
		return err
	}

	{
		api, err := c.API.replaceSecrets(c.Secrets)
		if err != nil {
			return err
		}

		c.API = api
	}

	for serviceName, service := range c.Services {
		for taskName, task := range service {
			sourceWithSecrets, err := task.Source.replaceSecrets(c.Secrets)
			if err != nil {
				return err
			}

			task.Source = sourceWithSecrets
			c.Services[serviceName][taskName] = task

			for destinationName, destination := range task.Destinations {
				destinationWithSecrets, err := destination.replaceSecrets(c.Secrets)
				if err != nil {
					return err
				}

				c.Services[serviceName][taskName].Destinations[destinationName] = destinationWithSecrets
			}
		}
	}

	return nil
}

func (c *Config) replaceSecretsNotifications() error {
	for notificationName, notification := range c.Definitions.Notifications {
		for name, value := range notification.Config {
			t, err := template.New("notification/config").Option("missingkey=error").Parse(value)
			if err != nil {
				return fmt.Errorf("parse notification config template: %w", err)
			}

			var tpl bytes.Buffer
			if err := t.Execute(&tpl, c.Secrets); err != nil {
				return fmt.Errorf("execute notification config template: %w", err)
			}

			notification.Config[name] = tpl.String()
		}

		c.Definitions.Notifications[notificationName] = notification
	}

	return nil
}

func (destination DriverDestinationConfig) replaceSecrets(secrets Secrets) (DriverDestinationConfig, error) {
	{
		configYaml, err := yaml.Marshal(destination.Config)
		if err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("marshal destination config: %w", err)
		}

		t, err := template.New("destination/config").Option("missingkey=error").Parse(string(configYaml))
		if err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("parse destination config template: %w", err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, secrets); err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("execute destination config template: %w", err)
		}

		var configInterface any
		if err := yaml.Unmarshal(tpl.Bytes(), &configInterface); err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("unmarshal destination config: %w", err)
		}

		destination.Config = configInterface
	}

	{
		envData, err := yaml.Marshal(destination.Env)
		if err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("marshal destination env: %w", err)
		}

		t, err := template.New("destination/env").Option("missingkey=error").Parse(string(envData))
		if err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("parse destination env template: %w", err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, secrets); err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("execute destination env template: %w", err)
		}

		var env map[string]string
		if err := yaml.Unmarshal(tpl.Bytes(), &env); err != nil {
			return DriverDestinationConfig{}, fmt.Errorf("unmarshal destination env: %w", err)
		}

		destination.Env = env
	}

	return destination, nil
}

func (source DriverSourceConfig) replaceSecrets(secrets Secrets) (DriverSourceConfig, error) {
	{
		configYaml, err := yaml.Marshal(source.Config)
		if err != nil {
			return DriverSourceConfig{}, fmt.Errorf("marshal source config: %w", err)
		}

		t, err := template.New("source/config").Option("missingkey=error").Parse(string(configYaml))
		if err != nil {
			return DriverSourceConfig{}, fmt.Errorf("parse source config template: %w", err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, secrets); err != nil {
			return DriverSourceConfig{}, fmt.Errorf("execute source config template: %w", err)
		}

		var configInterface any
		if err := yaml.Unmarshal(tpl.Bytes(), &configInterface); err != nil {
			return DriverSourceConfig{}, fmt.Errorf("unmarshal source config: %w", err)
		}

		source.Config = configInterface
	}

	{
		envData, err := yaml.Marshal(source.Env)
		if err != nil {
			return DriverSourceConfig{}, fmt.Errorf("marshal source env: %w", err)
		}

		t, err := template.New("source/env").Option("missingkey=error").Parse(string(envData))
		if err != nil {
			return DriverSourceConfig{}, fmt.Errorf("parse source env template: %w", err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, secrets); err != nil {
			return DriverSourceConfig{}, fmt.Errorf("execute source env template: %w", err)
		}

		var env map[string]string
		if err := yaml.Unmarshal(tpl.Bytes(), &env); err != nil {
			return DriverSourceConfig{}, fmt.Errorf("unmarshal source env: %w", err)
		}

		source.Env = env
	}

	return source, nil
}

func (api API) replaceSecrets(secrets Secrets) (API, error) {
	users := map[string]string{}

	for username, password := range api.Users {
		t, err := template.New("api/users").Option("missingkey=error").Parse(password)
		if err != nil {
			return API{}, fmt.Errorf("parse api users template: %w", err)
		}

		var tpl bytes.Buffer
		if err := t.Execute(&tpl, secrets); err != nil {
			return API{}, fmt.Errorf("execute api users template: %w", err)
		}

		users[username] = tpl.String()
	}

	return API{
		Users: users,
	}, nil
}
