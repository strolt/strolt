package config

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

func (fi *FileInfo) merge() (*Config, error) {
	overrideFi := fi
	override := overrideFi.Config

	for _, baseFi := range overrideFi.ExtendedFileInfoList {
		base, err := baseFi.merge()
		if err != nil {
			return nil, err
		}

		timezone, timeLocation, err := mergeTimeZone(base.TimeZone, override.TimeZone)
		if err != nil {
			return base, err
		}

		override.TimeZone = timezone
		override.timeLocation = timeLocation

		override.Tags = mergeTags(base.Tags, override.Tags)

		override.Secrets = mergeSecretsForConfig(base.Secrets, baseFi.ExtendedSecretsList, override.Secrets, overrideFi.ExtendedSecretsList)

		override.Definitions, err = mergeDefinitions(base.Definitions, override.Definitions)
		if err != nil {
			return base, errors.Wrapf(err, "cannot merge definitions from %s", overrideFi.ConfigPathname)
		}

		override.API, err = mergeAPI(base.API, override.API)
		if err != nil {
			return base, errors.Wrapf(err, "cannot merge api from %s", overrideFi.ConfigPathname)
		}

		override.Services, err = mergeServices(base.Services, override.Services)
		if err != nil {
			return base, errors.Wrapf(err, "cannot merge services from %s", overrideFi.ConfigPathname)
		}
	}

	return &override, nil
}

func mergeTimeZone(base string, override string) (string, *time.Location, error) {
	zone := base
	if override != "" {
		zone = override
	}

	timezone, err := time.LoadLocation(zone)
	if err != nil {
		return zone, timezone, fmt.Errorf("load time zone: %w", err)
	}

	return zone, timezone, nil
}

func uniqueStringSlice(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true

			list = append(list, entry)
		}
	}

	return list
}

func mergeTags(base []string, override []string) []string {
	tags := []string{}
	tags = append(tags, base...)
	tags = append(tags, override...)

	tags = uniqueStringSlice(tags)

	return tags
}

func mergeSecrets(base Secrets, override Secrets) Secrets {
	secrets := base

	maps.Copy(secrets, override)

	return secrets
}

func mergeSecretsForFile(base []Secrets, override Secrets) Secrets {
	secrets := Secrets{}

	for _, baseSecrets := range slices.Backward(base) {
		secrets = mergeSecrets(secrets, baseSecrets)
	}

	secrets = mergeSecrets(secrets, override)

	return secrets
}

func mergeSecretsForConfig(base Secrets, baseList []Secrets, override Secrets, overrideList []Secrets) Secrets {
	baseF := mergeSecretsForFile(baseList, base)
	overrideF := mergeSecretsForFile(overrideList, override)

	secrets := mergeSecrets(baseF, overrideF)

	return secrets
}

func mergeDefinitions(base Definitions, override Definitions) (Definitions, error) {
	destinations, err := mergeDefinitionsDestinations(base.Destinations, override.Destinations)
	if err != nil {
		return Definitions{}, err
	}

	notifications, err := mergeDefinitionsNotifications(base.Notifications, override.Notifications)
	if err != nil {
		return Definitions{}, err
	}

	return Definitions{
		Destinations:  destinations,
		Notifications: notifications,
	}, nil
}

func mergeDefinitionsDestinations(base map[string]DriverDestinationConfig, override map[string]DriverDestinationConfig) (map[string]DriverDestinationConfig, error) {
	definitions := base

	err := mergo.Merge(&definitions, override, mergo.WithOverride)
	if err != nil {
		return map[string]DriverDestinationConfig{}, fmt.Errorf("merge destination definitions: %w", err)
	}

	return definitions, nil
}

func mergeDefinitionsNotifications(base map[string]DriverNotificationConfig, override map[string]DriverNotificationConfig) (map[string]DriverNotificationConfig, error) {
	definitions := base

	err := mergo.Merge(&definitions, override, mergo.WithOverride)
	if err != nil {
		return map[string]DriverNotificationConfig{}, fmt.Errorf("merge notification definitions: %w", err)
	}

	return definitions, nil
}

func mergeServices(base map[string]Service, override map[string]Service) (map[string]Service, error) {
	services := base

	err := mergo.Merge(&services, override, mergo.WithOverride)
	if err != nil {
		return map[string]Service{}, fmt.Errorf("merge services: %w", err)
	}

	return services, nil
}

func (c *Config) mergeDestinationExtends() error {
	for serviceName, service := range c.Services {
		for taskName, task := range service {
			for destinationName, destination := range task.Destinations {
				mergedDestination, err := mergeDestinationExtends(c.Definitions.Destinations, destination)
				if err != nil {
					return err
				}

				c.Services[serviceName][taskName].Destinations[destinationName] = mergedDestination
			}
		}
	}

	return nil
}

func mergeDestinationExtends(mapDestinations map[string]DriverDestinationConfig, override DriverDestinationConfig) (DriverDestinationConfig, error) {
	if override.Extends == "" {
		return override, nil
	}

	destinationDefinition, ok := mapDestinations[override.Extends]
	if !ok {
		return override, errors.Errorf("destination '%s' not defined", override.Extends)
	}

	config, err := mergeDestinationExtendsConfig(override.Config, destinationDefinition.Config)
	if err != nil {
		return DriverDestinationConfig{}, err
	}

	override.Config = config

	if err := mergo.Merge(&override.Env, destinationDefinition.Env); err != nil {
		return DriverDestinationConfig{}, fmt.Errorf("merge destination env: %w", err)
	}

	if destinationDefinition.Driver != "" {
		override.Driver = destinationDefinition.Driver
	}

	override.Extends = ""

	return override, nil
}

func mergeDestinationExtendsConfig(base any, extends any) (any, error) {
	_base := map[string]any{
		"data": base,
	}

	_extends := map[string]any{
		"data": extends,
	}

	if err := mergo.Merge(&_base, _extends); err != nil {
		return base, fmt.Errorf("merge destination config: %w", err)
	}

	return _base["data"], nil
}

func mergeAPI(base API, override API) (API, error) {
	api := base

	err := mergo.Merge(&api, override, mergo.WithOverride)
	if err != nil {
		return API{}, fmt.Errorf("merge api: %w", err)
	}

	return api, nil
}
