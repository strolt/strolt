package config

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

func (c *Config) validate() error {
	if err := c.validateNotifications(); err != nil {
		return err
	}

	return c.validateServices()
}

func (c *Config) validateNotifications() error {
	for notificationName, notification := range c.Definitions.Notifications {
		if err := notification.validate(notificationName); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateServices() error {
	for serviceName, service := range c.Services {
		if len(service) == 0 {
			return fmt.Errorf("tasks must be exists service '%s'", serviceName)
		}

		for taskName, task := range service {
			if err := task.Source.validate(serviceName, taskName); err != nil {
				return err
			}

			if err := task.Schedule.validate(serviceName, taskName); err != nil {
				return err
			}

			for _, notificationName := range task.Notifications {
				_, ok := c.Definitions.Notifications[notificationName]
				if !ok {
					return fmt.Errorf("notification '%s' not defined", notificationName)
				}
			}

			if len(task.Destinations) == 0 {
				return fmt.Errorf("service '%s' task '%s' destinations must be fill", serviceName, taskName)
			}

			for destinationName, destination := range task.Destinations {
				if err := destination.validate(serviceName, taskName, destinationName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (schedule *Schedule) validate(serviceName string, taskName string) error {
	cr := cron.New()
	defer cr.Stop()

	if schedule.Backup != "" {
		if _, err := cr.AddFunc(schedule.Backup, func() {}); err != nil {
			return fmt.Errorf("invalid schedule backup in service '%s' task '%s' - %w", serviceName, taskName, err)
		}
	}

	if schedule.Prune != "" {
		if _, err := cr.AddFunc(schedule.Prune, func() {}); err != nil {
			return fmt.Errorf("invalid schedule prune in service '%s' task '%s' - %w", serviceName, taskName, err)
		}
	}

	return nil
}

func (source *DriverSourceConfig) validate(serviceName string, taskName string) error {
	if source.Driver == "" {
		return fmt.Errorf("service '%s' task '%s' source driver must be fill", serviceName, taskName)
	}

	if !dmanager.IsAvailableDriverSource(source.Driver) {
		return fmt.Errorf("service '%s' task '%s' source driver '%s' does not available: %s", serviceName, taskName, source.Driver, dmanager.GetAvailableDriverSource())
	}

	_, err := dmanager.GetSourceDriver(source.Driver, serviceName, taskName, source.Config, source.Env)
	if err != nil {
		return errors.Wrapf(err, "service '%s' task '%s'", serviceName, taskName)
	}

	return nil
}

func (destination *DriverDestinationConfig) validate(serviceName string, taskName string, destinationName string) error {
	if destination.Driver == "" {
		return fmt.Errorf("service '%s' task '%s' destination '%s' driver must be fill", serviceName, taskName, destinationName)
	}

	_, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, serviceName, taskName, destination.Config, destination.Env)
	if err != nil {
		return errors.Wrapf(err, "service '%s' task '%s' destination '%s'", serviceName, taskName, destinationName)
	}

	return nil
}

func (notification *DriverNotificationConfig) validate(notificationName string) error {
	if notification.Driver == "" {
		return fmt.Errorf("notification '%s' driver field is empty", notificationName)
	}

	if !dmanager.IsAvailableDriverNotification(notification.Driver) {
		return fmt.Errorf("notification '%s' driver '%s' does not available: %s", notificationName, notification.Driver, dmanager.GetAvailableDriverNotification())
	}

	for _, event := range notification.Events {
		if !sctxt.IsContextEventAvaliable(event) {
			return fmt.Errorf("notification '%s' event '%s' not available", notificationName, event)
		}
	}

	return nil
}
