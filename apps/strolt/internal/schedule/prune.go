package schedule

import (
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

func prune(serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)

	t, err := task.New(serviceName, taskName, sctxt.TSchedule, sctxt.OpTypePrune)
	if err != nil {
		log.Error(err)
	}

	defer t.Close()

	scheduleStart(t)

	for destinationName := range t.TaskConfig.Destinations {
		scheduleDestinationStart(t, destinationName)

		if _, err := t.DestinationPrune(destinationName, false); err != nil {
			scheduleDestinationError(t, destinationName, err)
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
