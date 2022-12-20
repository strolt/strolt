package task

import (
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t *Task) backupSourceToWorkDir() error {
	operation := ControllerOperation{
		ServiceName: t.ServiceName,
		TaskName:    t.TaskName,
		Operation:   COBackupSource,
	}

	if err := operation.Start(); err != nil {
		return err
	}
	defer operation.Stop()

	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return err
	}

	return sourceDriver.Backup(t.Context)
}

func (t *Task) Backup() error {
	var resultError error

	t.eventOperationStart()

	isSourceError := false
	isDestinationError := false

	{
		t.eventSourceStart()

		err := t.backupSourceToWorkDir()
		if err != nil {
			t.log.Error(err)
			t.eventSourceError(err)

			isSourceError = true
		} else {
			t.eventSourceStop()
		}
	}

	if !isSourceError {
		for destinationName := range t.TaskConfig.Destinations {
			t.eventDestinationStart(destinationName)

			backupOutput, err := t.backupWorkDirToDestination(destinationName)
			if err != nil {
				t.log.Error(err)
				t.eventDestinationError(destinationName, err)

				if !isDestinationError {
					isDestinationError = true
				}
			} else {
				t.eventDestinationStop(destinationName, backupOutput)
			}
		}
	}

	if isSourceError {
		resultError = fmt.Errorf("source: %s", t.Context.Source.Error)
		t.eventOperationError(resultError)
	}

	if isDestinationError {
		destinationErrors := []string{}

		for destinationName, operation := range t.Context.Destination {
			if operation.Error != "" {
				destinationErrors = append(destinationErrors, fmt.Sprintf("[%s]: %s", destinationName, operation.Error))
			}
		}

		resultError = fmt.Errorf("destination: %s", strings.Join(destinationErrors, ", "))
		t.eventOperationError(resultError)
	}

	if !isSourceError && !isDestinationError {
		t.eventOperationStop()
	}

	notificationWaitGroup.Wait()

	return resultError
}

func (t Task) backupWorkDirToDestination(destinationName string) (sctxt.BackupOutput, error) {
	operation := ControllerOperation{
		ServiceName:     t.ServiceName,
		TaskName:        t.TaskName,
		DestinationName: destinationName,
		Operation:       COBackupDestination,
	}

	if err := operation.Start(); err != nil {
		return sctxt.BackupOutput{}, err
	}
	defer operation.Stop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return sctxt.BackupOutput{}, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	return destinationDriver.Backup(t.Context)
}
