package schedule

import (
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

func scheduleStart(t *task.Task) {
	t.Context.Event = sctxt.EvOperationStart
	t.Context.Operation.Time.Start = time.Now()

	go sendNotifications(*t)
}

func scheduleStop(t *task.Task) {
	t.Context.Event = sctxt.EvOperationStop
	t.Context.Operation.Time.Stop = time.Now()

	go sendNotifications(*t)
}

func scheduleError(t *task.Task, err error) {
	t.Context.Event = sctxt.EvOperationError
	t.Context.Operation.Time.Stop = time.Now()
	t.Context.Operation.Error = err.Error()

	go sendNotifications(*t)
}

func scheduleSourceStart(t *task.Task) {
	t.Context.Event = sctxt.EvSourceStart
	t.Context.Source.Time.Start = time.Now()

	go sendNotifications(*t)
}

func scheduleSourceStop(t *task.Task) {
	t.Context.Event = sctxt.EvSourceStop
	t.Context.Source.Time.Stop = time.Now()

	go sendNotifications(*t)
}

func scheduleSourceError(t *task.Task, err error) {
	t.Context.Event = sctxt.EvSourceError
	t.Context.Source.Time.Stop = time.Now()
	t.Context.Source.Error = err.Error()

	go sendNotifications(*t)
}

func scheduleDestinationStart(t *task.Task, destinationName string) {
	operation, ok := t.Context.Destination[destinationName]
	if !ok {
		operation = context.Operation{}
	}

	t.Context.Event = sctxt.EvDestinationStart
	operation.Time.Start = time.Now()
	t.Context.Destination[destinationName] = operation

	go sendNotifications(*t)
}

func scheduleDestinationStop(t *task.Task, destinationName string) {
	operation, ok := t.Context.Destination[destinationName]
	if !ok {
		operation = context.Operation{}
	}

	t.Context.Event = sctxt.EvDestinationStop
	operation.Time.Stop = time.Now()
	t.Context.Destination[destinationName] = operation

	go sendNotifications(*t)
}

func scheduleDestinationError(t *task.Task, destinationName string, err error) {
	operation, ok := t.Context.Destination[destinationName]
	if !ok {
		operation = context.Operation{}
	}

	t.Context.Event = sctxt.EvDestinationError
	operation.Time.Stop = time.Now()
	operation.Error = err.Error()
	t.Context.Destination[destinationName] = operation

	go sendNotifications(*t)
}
