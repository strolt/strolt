package config

import (
	"fmt"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"gopkg.in/yaml.v3"
)

type TaskConfig struct {
	Schedule      Schedule
	Source        DriverSourceConfig
	Destinations  map[string]DriverDestinationConfig
	Notifications map[string]DriverNotificationConfig
	Tags          []string
}

func Get() Config {
	return config
}

func GetServiceNameList() []string {
	serviceNameList := []string{}

	for serviceName := range config.Services {
		serviceNameList = append(serviceNameList, serviceName)
	}

	return serviceNameList
}

func GetTaskNameList(serviceName string) []string {
	taskNameList := []string{}

	for taskName := range config.Services[serviceName] {
		taskNameList = append(taskNameList, taskName)
	}

	return taskNameList
}

func (notification *DriverNotificationConfig) IsAvailableEvent(event sctxt.EventType) bool {
	for _, e := range notification.Events {
		if e == event {
			return true
		}
	}

	return false
}

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

func GetLocation() *time.Location {
	l := *config.timeLocation
	return &l
}

func GetConfigForTask(serviceName string, taskName string) (TaskConfig, error) {
	cTask := TaskConfig{}
	c := Get()

	task, ok := c.Services[serviceName][taskName]
	if !ok {
		return cTask, fmt.Errorf("not found config for task %s", taskName)
	}

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

func Yaml() string {
	data, _ := yaml.Marshal(config)

	return string(data)
}

type FileLists struct {
	SecretsFileList        []string
	ExtendedConfigFileList []string
}
