package task

import "github.com/strolt/strolt/apps/strolt/internal/dmanager"

func (t Task) SendNotifications() error {
	for _, notification := range t.TaskConfig.Notifications {
		if notification.IsAvailableEvent(t.Context.Event) {
			driverNotification, err := dmanager.GetNotificationDriver(notification.Driver, t.ServiceName, t.TaskName, notification.Config)
			if err != nil {
				return err
			}

			driverNotification.Send(t.Context)
		}
	}

	return nil
}
