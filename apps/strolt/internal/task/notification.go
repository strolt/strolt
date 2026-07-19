package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

// notifyWaitGroup returns this task's notification WaitGroup, creating it on
// first use. It is only ever accessed from the task's own operation goroutine
// (events are emitted sequentially), so the lazy initialization needs no lock.
func (t *Task) notifyWaitGroup() *sync.WaitGroup {
	if t.notificationWaitGroup == nil {
		t.notificationWaitGroup = &sync.WaitGroup{}
	}

	return t.notificationWaitGroup
}

// SendNotifications sends the current context event to all matching notification drivers.
func (t *Task) SendNotifications() error {
	for _, notification := range t.TaskConfig.Notifications {
		if notification.IsAvailableEvent(t.Context.Event) {
			driverNotification, err := dmanager.GetNotificationDriver(notification.Driver, t.ServiceName, t.TaskName, notification.Config)
			if err != nil {
				return fmt.Errorf("get notification driver: %w", err)
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

	t.notifyWaitGroup().Go(func() {
		if err := tCopy.SendNotifications(); err != nil {
			log := logger.New()
			log.Warnf("send notification error: %s", err)
		}
	})
}

func (t *Task) eventOperationStart() {
	t.Context.Event = sctxt.EvOperationStart
	t.Context.Operation.Time.Start = time.Now()

	t.sendNotifications()
}

func (t *Task) eventOperationStop() {
	t.Context.Event = sctxt.EvOperationStop
	t.Context.Operation.Time.Stop = time.Now()

	t.UpdateMetricsAfterTaskFinish()

	t.sendNotifications()
}

func (t *Task) eventOperationError(err error) {
	t.Context.Event = sctxt.EvOperationError
	t.Context.Operation.Time.Stop = time.Now()
	t.Context.Operation.Error = err.Error()

	t.UpdateMetricsAfterTaskFinish()

	t.sendNotifications()
}

func (t *Task) eventSourceStart() {
	t.Context.Event = sctxt.EvSourceStart
	t.Context.Source.Time.Start = time.Now()

	t.sendNotifications()
}

func (t *Task) eventSourceError(err error) {
	t.Context.Event = sctxt.EvSourceError
	t.Context.Source.Error = err.Error()
	t.Context.Source.Time.Stop = time.Now()

	t.sendNotifications()
}

func (t *Task) eventSourceStop() {
	t.Context.Event = sctxt.EvSourceStop
	t.Context.Source.Time.Stop = time.Now()

	t.sendNotifications()
}

func (t *Task) eventDestinationStart(destinationName string) {
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Start = time.Now()

	t.Context.Event = sctxt.EvDestinationStart
	t.Context.Destination[destinationName] = operation

	t.sendNotifications()
}

func (t *Task) eventDestinationError(destinationName string, err error) {
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Stop = time.Now()
	operation.Error = err.Error()

	t.Context.Event = sctxt.EvDestinationError
	t.Context.Destination[destinationName] = operation

	t.sendNotifications()
}

func (t *Task) eventDestinationStop(destinationName string, output sctxt.BackupOutput) {
	operation, ok := t.Context.Destination[destinationName]

	if !ok {
		operation = context.DestinationOperation{}
	}

	operation.Time.Stop = time.Now()
	operation.BackupOutput = output

	t.Context.Event = sctxt.EvDestinationStop
	t.Context.Destination[destinationName] = operation

	t.sendNotifications()
}
