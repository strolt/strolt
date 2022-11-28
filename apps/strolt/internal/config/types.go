package config

import (
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

type Secrets map[string]string

type Schedule struct {
	Backup string `yaml:"backup,omitempty"`
	Prune  string `yaml:"prune,omitempty"`
}

type DriverSourceConfig struct {
	Driver dmanager.Source   `yaml:"driver,omitempty"`
	Config interface{}       `yaml:"config,omitempty"`
	Env    map[string]string `yaml:"env,omitempty"`
}

type DriverDestinationConfig struct {
	Extends string               `yaml:"extends,omitempty"`
	Driver  dmanager.Destination `yaml:"driver,omitempty"`
	Config  interface{}          `yaml:"config,omitempty"`
	Env     map[string]string    `yaml:"env,omitempty"`
}

type DriverNotificationConfig struct {
	Driver dmanager.Notification `yaml:"driver,omitempty"`
	Config map[string]string     `yaml:"config,omitempty"`
	Events []sctxt.EventType     `yaml:"events,omitempty"`
}

type Task struct {
	Source        DriverSourceConfig                 `yaml:"source,omitempty"`
	Destinations  map[string]DriverDestinationConfig `yaml:"destinations,omitempty"`
	Notifications []string                           `yaml:"notifications,omitempty"`
	Schedule      Schedule                           `yaml:"schedule,omitempty"`
	Tags          []string                           `yaml:"tags,omitempty"`
}

type Service map[string]Task

type Extends struct {
	Secrets []string `yaml:"secrets,omitempty"`
	Configs []string `yaml:"configs,omitempty"`
}

type Definitions struct {
	Destinations  map[string]DriverDestinationConfig  `yaml:"destinations,omitempty"`
	Notifications map[string]DriverNotificationConfig `yaml:"notifications,omitempty"`
}

type Config struct {
	TimeZone            string             `yaml:"timezone,omitempty"`
	timeLocation        *time.Location     `json:"-" yaml:"-"`
	DisableWatchChanges bool               `yaml:"disableWatchChanges,omitempty"` // TODO: move to ENV
	Services            map[string]Service `yaml:"services"`
	Tags                []string           `yaml:"tags,omitempty"`
	Secrets             Secrets            `yaml:"secrets,omitempty"`
	Extends             Extends            `yaml:"extends,omitempty"`
	Definitions         Definitions        `yaml:"definitions,omitempty"`
}

var (
	config        = Config{}
	fileList      = []string{}
	initialConfig = Config{
		timeLocation: time.Local,
		TimeZone:     time.Local.String(),
	}
)
