package schedule

import (
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

func backup(serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)
	t, err := task.New(serviceName, taskName, sctxt.TSchedule, sctxt.OpTypeBackup)

	if err != nil {
		log.Error(err)
	}

	defer t.Close()

	scheduleStart(t)
	{
		scheduleSourceStart(t)
		if err := t.BackupSourceToWorkDir(); err != nil {
			log.Error(err)
			scheduleSourceError(t, err)
			scheduleError(t, fmt.Errorf("source: %w", err))

			return
		}
		scheduleSourceStop(t)
	}

	for destinationName := range t.TaskConfig.Destinations {
		scheduleDestinationStart(t, destinationName)

		backupOutput, err := t.BackupWorkDirToDestination(destinationName)
		if err != nil {
			scheduleDestinationError(t, destinationName, err)
		}

		t.Context.Destination[destinationName] = context.Operation{
			Time:  t.Context.Destination[destinationName].Time,
			Error: t.Context.Destination[destinationName].Error,

			BackupOutput: backupOutput,
		}

		scheduleDestinationStop(t, destinationName)
	}

	var destinationErrors []string

	for destinationName, operation := range t.Context.Destination {
		if operation.Error != "" {
			destinationErrors = append(destinationErrors, fmt.Sprintf("[%s]: %s", destinationName, operation.Error))
		}
	}

	if len(destinationErrors) == 0 {
		scheduleStop(t)
	} else {
		scheduleError(t, fmt.Errorf("destination: %s", strings.Join(destinationErrors, ", ")))
	}
}
