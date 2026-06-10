package config

import (
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// OperationMode defines how data is transferred between source and destination.
type OperationMode string

// Supported operation modes.
const (
	OperationModeCopy       OperationMode = "copy"
	OperationModePreferPipe OperationMode = "prefer-pipe"
	OperationModePipe       OperationMode = "pipe"
)

// Secrets maps secret names to their values.
type Secrets map[string]string

// Schedule defines cron expressions for backup and prune operations.
type Schedule struct {
	Backup string `yaml:"backup,omitempty"`
	Prune  string `yaml:"prune,omitempty"`
}

// DriverSourceConfig describes a source driver and its settings.
type DriverSourceConfig struct {
	Driver dmanager.Source   `yaml:"driver,omitempty"`
	Config any               `yaml:"config,omitempty"`
	Env    map[string]string `yaml:"env,omitempty"`
}

// DriverDestinationConfig describes a destination driver and its settings.
type DriverDestinationConfig struct {
	Extends string               `yaml:"extends,omitempty"`
	Driver  dmanager.Destination `yaml:"driver,omitempty"`
	Config  any                  `yaml:"config,omitempty"`
	Env     map[string]string    `yaml:"env,omitempty"`
}

// DriverNotificationConfig describes a notification driver and its settings.
type DriverNotificationConfig struct {
	Driver dmanager.Notification `yaml:"driver,omitempty"`
	Config map[string]string     `yaml:"config,omitempty"`
	Events []sctxt.EventType     `yaml:"events,omitempty"`
}

// Task describes a single backup task configuration.
type Task struct {
	OperationMode OperationMode                      `yaml:"operationMode,omitempty"`
	Source        DriverSourceConfig                 `yaml:"source,omitempty"`
	Destinations  map[string]DriverDestinationConfig `yaml:"destinations,omitempty"`
	Notifications []string                           `yaml:"notifications,omitempty"`
	Schedule      Schedule                           `yaml:"schedule,omitempty"`
	Tags          []string                           `yaml:"tags,omitempty"`
}

// Service maps task names to their configurations.
type Service map[string]Task

// Extends lists external config and secrets files to include.
type Extends struct {
	Secrets []string `yaml:"secrets,omitempty"`
	Configs []string `yaml:"configs,omitempty"`
}

// Definitions holds reusable destination and notification configurations.
type Definitions struct {
	Destinations  map[string]DriverDestinationConfig  `yaml:"destinations,omitempty"`
	Notifications map[string]DriverNotificationConfig `yaml:"notifications,omitempty"`
}

// Config is the root strolt configuration.
type Config struct {
	TimeZone     string             `yaml:"timezone,omitempty"`
	timeLocation *time.Location     `json:"-"                     yaml:"-"`
	Services     map[string]Service `yaml:"services"`
	Tags         []string           `yaml:"tags,omitempty"`
	Secrets      Secrets            `yaml:"secrets,omitempty"`
	Extends      Extends            `yaml:"extends,omitempty"`
	Definitions  Definitions        `yaml:"definitions,omitempty"`
	API          API                `yaml:"api,omitempty"`
}

// API holds API server configuration such as user credentials.
type API struct {
	Users map[string]string `yaml:"users,omitempty"`
}

var (
	config   = Config{}
	fileList = []string{}
)
