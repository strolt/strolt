package task

import (
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

var notificationWaitGroup = sync.WaitGroup{}

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

func (t *Task) sendNotifications() {
	if t.isNotificationsDisabled {
		return
	}

	tCopy, err := t.Clone()
	if err != nil {
		logger.New().Error("error copy task")
		return
	}

	notificationWaitGroup.Add(1)

	go func() {
		if err := tCopy.SendNotifications(); err != nil {
			log := logger.New()
			log.Warnf("send notification error: %s", err)
		}

		notificationWaitGroup.Done()
	}()
}

func (t *Task) eventOperationStart() {
	t.Lock()
	t.Context.Event = sctxt.EvOperationStart
	t.Context.Operation.Time.Start = time.Now()
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventOperationStop() {
	t.Lock()
	t.Context.Event = sctxt.EvOperationStop
	t.Context.Operation.Time.Stop = time.Now()
	t.Unlock()

	t.UpdateMetricsAfterTaskFinish()

	t.sendNotifications()
}

func (t *Task) eventOperationError(err error) {
	t.Lock()
	t.Context.Event = sctxt.EvOperationError
	t.Context.Operation.Time.Stop = time.Now()
	t.Context.Operation.Error = err.Error()
	t.Unlock()

	t.UpdateMetricsAfterTaskFinish()

	t.sendNotifications()
}

func (t *Task) eventSourceStart() {
	t.Lock()
	t.Context.Event = sctxt.EvSourceStart
	t.Context.Source.Time.Start = time.Now()
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventSourceError(err error) {
	t.Lock()
	t.Context.Event = sctxt.EvSourceError
	t.Context.Source.Error = err.Error()
	t.Context.Source.Time.Stop = time.Now()
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventSourceStop() {
	t.Lock()
	t.Context.Event = sctxt.EvSourceStop
	t.Context.Source.Time.Stop = time.Now()
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventDestinationStart(destinationName string) {
	t.Lock()
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Start = time.Now()

	t.Context.Event = sctxt.EvDestinationStart
	t.Context.Destination[destinationName] = operation
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventDestinationError(destinationName string, err error) {
	t.Lock()
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Stop = time.Now()
	operation.Error = err.Error()

	t.Context.Event = sctxt.EvDestinationError
	t.Context.Destination[destinationName] = operation
	t.Unlock()

	t.sendNotifications()
}

func (t *Task) eventDestinationStop(destinationName string, output sctxt.BackupOutput) {
	t.Lock()
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Stop = time.Now()
	operation.BackupOutput = output

	t.Context.Event = sctxt.EvDestinationStop
	t.Context.Destination[destinationName] = operation
	t.Unlock()

	t.sendNotifications()
}
