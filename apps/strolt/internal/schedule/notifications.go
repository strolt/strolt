package schedule

import (
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

func sendNotifications(t task.Task) {
	if len(t.TaskConfig.Notifications) != 0 {
		for _, notification := range t.TaskConfig.Notifications {
			if notification.IsAvailableEvent(t.Context.Event) {
				driverNotification, err := dmanager.GetNotificationDriver(notification.Driver, t.ServiceName, t.TaskName, notification.Config)
				if err == nil {
					go driverNotification.Send(t.Context)
				}
			}
		}
	}
}
