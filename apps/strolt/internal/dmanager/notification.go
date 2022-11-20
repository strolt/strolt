package dmanager

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/console"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/slack"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/telegram"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

type Notification string

const (
	DriverNotificationConsole  Notification = "console"
	DriverNotificationEmail    Notification = "email"
	DriverNotificationSlack    Notification = "slack"
	DriverNotificationTelegram Notification = "telegram"
)

func GetAvailableDriverNotification() []Notification {
	return []Notification{
		DriverNotificationConsole,
		DriverNotificationEmail,
		DriverNotificationSlack,
		DriverNotificationTelegram,
	}
}

func IsAvailableDriverNotification(driver Notification) bool {
	for _, d := range GetAvailableDriverNotification() {
		if driver == d {
			return true
		}
	}

	return false
}

func GetNotificationDriver(driver Notification, serviceName string, taskName string, driverConfig interface{}) (interfaces.DriverNotificationInterface, error) {
	notificationDrivers := map[Notification]interfaces.DriverNotificationInterface{
		DriverNotificationConsole:  console.New(),
		DriverNotificationSlack:    slack.New(slack.Params{}),
		DriverNotificationTelegram: telegram.New(telegram.Params{}),
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
		return nil, err
	}

	return d, nil
}
