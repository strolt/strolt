package dmanager

import (
	"fmt"
	"slices"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/console"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/slack"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/telegram"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/webhook"
	"github.com/strolt/strolt/shared/logger"
)

// Notification is the name of a notification driver.
type Notification string

// Supported notification drivers.
const (
	DriverNotificationConsole  Notification = "console"
	DriverNotificationSlack    Notification = "slack"
	DriverNotificationTelegram Notification = "telegram"
	DriverNotificationWebhook  Notification = "webhook"
)

// GetAvailableDriverNotification returns the supported notification drivers.
func GetAvailableDriverNotification() []Notification {
	return []Notification{
		DriverNotificationConsole,
		DriverNotificationSlack,
		DriverNotificationTelegram,
		DriverNotificationWebhook,
	}
}

// IsAvailableDriverNotification reports whether the notification driver is supported.
func IsAvailableDriverNotification(driver Notification) bool {
	return slices.Contains(GetAvailableDriverNotification(), driver)
}

// GetNotificationDriver creates and configures the requested notification driver.
func GetNotificationDriver(driver Notification, serviceName string, taskName string, driverConfig any) (interfaces.DriverNotificationInterface, error) {
	notificationDrivers := map[Notification]interfaces.DriverNotificationInterface{
		DriverNotificationConsole:  console.New(),
		DriverNotificationSlack:    slack.New(slack.Params{}),
		DriverNotificationTelegram: telegram.New(telegram.Params{}),
		DriverNotificationWebhook:  webhook.New(),
	}

	d, ok := notificationDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("notification driver '%s' does not exists", driver)
	}

	loggerFields := logger.Fields{
		"driverType":  "notification",
		"serviceName": serviceName,
		"taskName":    taskName,
		"driver":      driver,
	}
	d.SetLogger(logger.New().WithFields(loggerFields))

	if err := d.SetConfig(driverConfig); err != nil {
		return nil, fmt.Errorf("set notification driver config: %w", err)
	}

	return d, nil
}
