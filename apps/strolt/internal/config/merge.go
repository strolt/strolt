package config

import (
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

		override.DisableWatchChanges = mergeDisableWatchChanges(base.DisableWatchChanges, override.DisableWatchChanges)
		override.Tags = mergeTags(base.Tags, override.Tags)

		override.Secrets = mergeSecretsForConfig(base.Secrets, baseFi.ExtendedSecretsList, override.Secrets, overrideFi.ExtendedSecretsList)

		override.Definitions, err = mergeDefinitions(base.Definitions, override.Definitions)
		if err != nil {
			return base, errors.Wrapf(err, "cannot merge definitions from %s", overrideFi.ConfigPathname)
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

	return zone, timezone, err
}

func mergeDisableWatchChanges(base bool, override bool) bool {
	if override {
		return override
	}

	return base
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

	for key, value := range override {
		secrets[key] = value
	}

	return secrets
}

func mergeSecretsForFile(base []Secrets, override Secrets) Secrets {
	secrets := Secrets{}

	for i := len(base) - 1; i >= 0; i-- {
		secrets = mergeSecrets(secrets, base[i])
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
		return map[string]DriverDestinationConfig{}, err
	}

	return definitions, nil
}

func mergeDefinitionsNotifications(base map[string]DriverNotificationConfig, override map[string]DriverNotificationConfig) (map[string]DriverNotificationConfig, error) {
	definitions := base

	err := mergo.Merge(&definitions, override, mergo.WithOverride)
	if err != nil {
		return map[string]DriverNotificationConfig{}, err
	}

	return definitions, nil
}

func mergeServices(base map[string]Service, override map[string]Service) (map[string]Service, error) {
	services := base

	err := mergo.Merge(&services, override, mergo.WithOverride)
	if err != nil {
		return map[string]Service{}, err
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
		return DriverDestinationConfig{}, err
	}

	if destinationDefinition.Driver != "" {
		override.Driver = destinationDefinition.Driver
	}

	override.Extends = ""

	return override, nil
}

func mergeDestinationExtendsConfig(base interface{}, extends interface{}) (interface{}, error) {
	_base := map[string]interface{}{
		"data": base,
	}

	_extends := map[string]interface{}{
		"data": extends,
	}

	if err := mergo.Merge(&_base, _extends); err != nil {
		return base, err
	}

	return _base["data"], nil
}
