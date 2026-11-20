package config

import (
	"strings"
	"testing"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// validConfig returns a minimal configuration passing validation; individual
// tests mutate it to trigger specific validation errors.
func validConfig() Config {
	return Config{
		TimeZone: "", timeLocation: nil, Tags: nil, Secrets: nil, Extends: Extends{Secrets: nil, Configs: nil}, Definitions: Definitions{Destinations: nil, Notifications: nil}, API: API{Users: nil},
		Services: map[string]Service{
			"svc": {
				"task": {
					OperationMode: "", Notifications: nil, Schedule: Schedule{Backup: "", Prune: ""}, Tags: nil,
					Source: DriverSourceConfig{
						Env:    nil,
						Driver: dmanager.DriverSourceLocal,
						Config: map[string]any{"path": "."},
					},
					Destinations: map[string]DriverDestinationConfig{
						"dest": {
							Extends: "", Env: nil,
							Driver: dmanager.DriverDestinationLocal,
							Config: map[string]any{"path": "."},
						},
					},
				},
			},
		},
	}
}

func TestValidateAcceptsMinimalConfig(t *testing.T) {
	cfg := validConfig()

	if err := cfg.validate(t.Context()); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "service without tasks",
			mutate: func(c *Config) {
				c.Services["empty"] = Service{}
			},
			wantErr: "tasks must be exists service 'empty'",
		},
		{
			name: "source driver missing",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Source = DriverSourceConfig{Driver: "", Config: nil, Env: nil}
				c.Services["svc"]["task"] = task
			},
			wantErr: "source driver must be fill",
		},
		{
			name: "unknown source driver",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Source.Driver = dmanager.Source("nope")
				c.Services["svc"]["task"] = task
			},
			wantErr: "does not available",
		},
		{
			name: "destinations empty",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Destinations = nil
				c.Services["svc"]["task"] = task
			},
			wantErr: "destinations must be fill",
		},
		{
			name: "destination driver missing",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Destinations["dest"] = DriverDestinationConfig{Extends: "", Driver: "", Config: nil, Env: nil}
				c.Services["svc"]["task"] = task
			},
			wantErr: "driver must be fill",
		},
		{
			name: "invalid backup cron",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Schedule.Backup = "not a cron"
				c.Services["svc"]["task"] = task
			},
			wantErr: "invalid schedule backup",
		},
		{
			name: "invalid prune cron",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Schedule.Prune = "not a cron"
				c.Services["svc"]["task"] = task
			},
			wantErr: "invalid schedule prune",
		},
		{
			name: "undefined notification reference",
			mutate: func(c *Config) {
				task := c.Services["svc"]["task"]
				task.Notifications = []string{"ghost"}
				c.Services["svc"]["task"] = task
			},
			wantErr: "notification 'ghost' not defined",
		},
		{
			name: "notification driver empty",
			mutate: func(c *Config) {
				c.Definitions.Notifications = map[string]DriverNotificationConfig{
					"note": {Config: nil, Driver: "", Events: []sctxt.EventType{sctxt.EvOperationStop}},
				}
				task := c.Services["svc"]["task"]
				task.Notifications = []string{"note"}
				c.Services["svc"]["task"] = task
			},
			wantErr: "driver field is empty",
		},
		{
			name: "unknown notification driver",
			mutate: func(c *Config) {
				c.Definitions.Notifications = map[string]DriverNotificationConfig{
					"note": {Config: nil, Events: nil, Driver: dmanager.Notification("sms")},
				}
				task := c.Services["svc"]["task"]
				task.Notifications = []string{"note"}
				c.Services["svc"]["task"] = task
			},
			wantErr: "does not available",
		},
		{
			name: "unknown notification event",
			mutate: func(c *Config) {
				c.Definitions.Notifications = map[string]DriverNotificationConfig{
					"note": {
						Config: nil,
						Driver: dmanager.DriverNotificationConsole,
						Events: []sctxt.EventType{sctxt.EventType("SOMETHING")},
					},
				}
				task := c.Services["svc"]["task"]
				task.Notifications = []string{"note"}
				c.Services["svc"]["task"] = task
			},
			wantErr: "not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(&cfg)

			err := cfg.validate(t.Context())
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}
