package api

import (
	"fmt"
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/apiu"
)

// Config is the API representation of the strolt configuration.
type Config struct {
	TimeZone            string                   `json:"timezone"`
	DisableWatchChanges bool                     `json:"disableWatchChanges"`
	Tags                []string                 `json:"tags"`
	Services            map[string]ConfigService `json:"services"`
} //	@name	Config

// ConfigService maps task names to their task configuration.
type ConfigService map[string]ConfigServiceTask //	@name	ConfigService

// ConfigServiceTask describes a single task configuration.
type ConfigServiceTask struct {
	Source        ConfigServiceTaskSource                 `json:"source"`
	Destinations  map[string]ConfigServiceTaskDestination `json:"destinations"`
	Notifications []ConfigServiceTaskNotification         `json:"notifications"`
	Schedule      ConfigServiceTaskSchedule               `json:"schedule"`
	Tags          []string                                `json:"tags"`
} //	@name	ConfigServiceTask

// ConfigServiceTaskSource describes the source driver of a task.
type ConfigServiceTaskSource struct {
	Driver string `json:"driver"`
} //	@name	ConfigServiceTaskSource

// ConfigServiceTaskSchedule describes the backup and prune schedules of a task.
type ConfigServiceTaskSchedule struct {
	Backup string `json:"backup"`
	Prune  string `json:"prune"`
} //	@name	ConfigServiceTaskSchedule

// ConfigServiceTaskDestination describes a destination driver of a task.
type ConfigServiceTaskDestination struct {
	Driver string `json:"driver"`
} //	@name	ConfigServiceTaskDestination

// ConfigServiceTaskNotification describes a notification driver of a task.
type ConfigServiceTaskNotification struct {
	Driver string            `json:"driver"`
	Name   string            `json:"name"`
	Events []sctxt.EventType `enums:"OPERATION_START,OPERATION_STOP,OPERATION_ERROR,SOURCE_START,SOURCE_STOP,SOURCE_ERROR,DESTINATION_START,DESTINATION_STOP,DESTINATION_ERROR" json:"events"`
} //	@name	ConfigServiceTaskNotification

// getConfig godoc
//
//	@Id			getConfig
//	@Summary	Show config
//	@Security	BasicAuth
//	@success	200	{object}	Config
//	@Router		/api/v1/config [get].
func (api *API) getConfig(w http.ResponseWriter, r *http.Request) {
	c := config.Get()

	tags := make([]string, 0, len(c.Tags))
	tags = append(tags, c.Tags...)

	services := map[string]ConfigService{}

	for serviceName, service := range c.Services {
		s := ConfigService{}

		for taskName, task := range service {
			notifications := []ConfigServiceTaskNotification{}

			for _, notificationName := range task.Notifications {
				notificationDefinition, ok := c.Definitions.Notifications[notificationName]
				if !ok {
					apiu.RenderJSON500(w, r, fmt.Errorf("not found notification definition '%s'", notificationName))
					return
				}

				notifications = append(notifications, ConfigServiceTaskNotification{
					Name:   notificationName,
					Driver: string(notificationDefinition.Driver),
					Events: notificationDefinition.Events,
				})
			}

			tags := make([]string, 0, len(task.Tags))
			tags = append(tags, task.Tags...)

			destinations := map[string]ConfigServiceTaskDestination{}

			for destinationName, destination := range task.Destinations {
				destinations[destinationName] = ConfigServiceTaskDestination{
					Driver: string(destination.Driver),
				}
			}

			t := ConfigServiceTask{
				Source: ConfigServiceTaskSource{
					Driver: string(task.Source.Driver),
				},
				Schedule: ConfigServiceTaskSchedule{
					Backup: task.Schedule.Backup,
					Prune:  task.Schedule.Prune,
				},
				Notifications: notifications,
				Tags:          tags,
				Destinations:  destinations,
			}

			s[taskName] = t
		}

		services[serviceName] = s
	}

	apiu.RenderJSON200(w, r, Config{
		TimeZone:            c.TimeZone,
		DisableWatchChanges: env.IsWatchFilesDisabled(),
		Tags:                tags,
		Services:            services,
	})
}
