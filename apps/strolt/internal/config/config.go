// Package config loads, merges, and validates the strolt configuration.
package config

import (
	"fmt"
	"slices"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"gopkg.in/yaml.v3"
)

// TaskConfig is the resolved configuration for a single task.
type TaskConfig struct {
	Schedule      Schedule
	Source        DriverSourceConfig
	OperationMode OperationMode
	Destinations  map[string]DriverDestinationConfig
	Notifications map[string]DriverNotificationConfig
	Tags          []string
}

// Get returns the currently loaded configuration.
func Get() Config {
	return config
}

// GetServiceNameList returns the names of all configured services.
func GetServiceNameList() []string {
	serviceNameList := make([]string, 0, len(config.Services))

	for serviceName := range config.Services {
		serviceNameList = append(serviceNameList, serviceName)
	}

	return serviceNameList
}

// GetTaskNameList returns the task names configured for the given service.
func GetTaskNameList(serviceName string) []string {
	taskNameList := make([]string, 0, len(config.Services[serviceName]))

	for taskName := range config.Services[serviceName] {
		taskNameList = append(taskNameList, taskName)
	}

	return taskNameList
}

// IsAvailableEvent reports whether the notification is enabled for the given event.
func (notification *DriverNotificationConfig) IsAvailableEvent(event sctxt.EventType) bool {
	return slices.Contains(notification.Events, event)
}

// GetDestinationNameList returns the destination names configured for the given task.
func GetDestinationNameList(serviceName string, taskName string) []string {
	task, ok := config.Services[serviceName][taskName]
	if !ok {
		return []string{}
	}

	destinationNameList := make([]string, len(task.Destinations))

	i := 0

	for destinationName := range task.Destinations {
		destinationNameList[i] = destinationName
		i++
	}

	return destinationNameList
}

// GetLocation returns a copy of the configured time location.
func GetLocation() *time.Location {
	l := *config.timeLocation
	return &l
}

// GetConfigForTask returns the resolved configuration for the given task.
func GetConfigForTask(serviceName string, taskName string) (TaskConfig, error) {
	cTask := TaskConfig{}
	c := Get()

	task, ok := c.Services[serviceName][taskName]
	if !ok {
		return cTask, fmt.Errorf("not found config for task %s", taskName)
	}

	cTask.OperationMode = task.OperationMode

	cTask.Schedule = task.Schedule

	cTask.Tags = append(cTask.Tags, append(c.Tags, task.Tags...)...)

	cTask.Source = c.Services[serviceName][taskName].Source

	cTask.Destinations = task.Destinations

	notifications := map[string]DriverNotificationConfig{}
	for _, notificationName := range task.Notifications {
		notifications[notificationName] = c.Definitions.Notifications[notificationName]
	}

	cTask.Notifications = notifications

	return cTask, nil
}

// Yaml returns the loaded configuration rendered as YAML.
func Yaml() string {
	data, _ := yaml.Marshal(config)

	return string(data)
}

// FileLists groups the secrets and extended config file paths.
type FileLists struct {
	SecretsFileList        []string
	ExtendedConfigFileList []string
}
