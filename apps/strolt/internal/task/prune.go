package task

import (
	"errors"
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t *Task) prune(destinationName string, isDryRun bool) ([]interfaces.Snapshot, error) {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return []interfaces.Snapshot{}, errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return []interfaces.Snapshot{}, err
	}

	return destinationDriver.Prune(t.Context, isDryRun)
}

func (t *Task) Prune(destinationName string, isDryRun bool) ([]interfaces.Snapshot, error) {
	if err := t.managerStart(sctxt.OpTypePrune); err != nil {
		return []interfaces.Snapshot{}, err
	}
	defer t.managerStop()

	if isDryRun {
		t.isNotificationsDisabled = true
	}

	t.eventOperationStart()
	t.eventDestinationStart(destinationName)

	snapshotList, err := t.prune(destinationName, isDryRun)
	if err != nil {
		t.eventDestinationError(destinationName, err)
		t.eventOperationError(err)
	} else {
		t.eventDestinationStop(destinationName, sctxt.BackupOutput{})
		t.eventOperationStop()
	}

	notificationWaitGroup.Wait()

	return snapshotList, err
}

func (t *Task) PruneAll() error {
	var resultError error

	t.eventOperationStart()

	isPruneError := false

	for destinationName := range t.TaskConfig.Destinations {
		t.eventDestinationStart(destinationName)

		_, err := t.prune(destinationName, false)
		if err != nil {
			if !isPruneError {
				isPruneError = true
			}

			t.eventDestinationError(destinationName, err)
		} else {
			t.eventDestinationStop(destinationName, sctxt.BackupOutput{})
		}
	}

	if isPruneError {
		destinationErrors := []string{}

		for destinationName, operation := range t.Context.Destination {
			if operation.Error != "" {
				destinationErrors = append(destinationErrors, fmt.Sprintf("[%s]: %s", destinationName, operation.Error))
			}
		}

		resultError = fmt.Errorf("destination: %s", strings.Join(destinationErrors, ", "))

		t.eventOperationError(resultError)
	} else {
		t.eventOperationStop()
	}

	notificationWaitGroup.Wait()

	return resultError
}
