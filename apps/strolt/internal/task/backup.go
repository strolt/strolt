package task

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

var (
	ErrNotSupportedPipeMode = errors.New("source or one of destinations does not support pipe mode")
)

func (t *Task) backupSourceToWorkDir() error {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return err
	}

	return sourceDriver.Backup(t.Context)
}

func (t *Task) getSourceDriver() (interfaces.DriverSourceInterface, error) {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return nil, err
	}

	return sourceDriver, nil
}

func (t *Task) isAvailableBackupPipe() (bool, error) {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return false, err
	}

	if !sourceDriver.IsSupportedBackupPipe(t.Context) {
		return false, nil
	}

	for destinationName := range t.TaskConfig.Destinations {
		destinationDriver, err := t.getDestinationDriver(destinationName)
		if err != nil {
			return false, err
		}

		if !destinationDriver.IsSupportedBackupPipe(t.Context) {
			return false, nil
		}
	}

	return true, nil
}

func (t *Task) backupManual() error {
	if err := t.managerStart(sctxt.OpTypeBackup); err != nil {
		return err
	}
	defer t.managerStop()

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

func (t *Task) backupPipe() error {
	if err := t.managerStart(sctxt.OpTypeBackup); err != nil {
		return err
	}
	defer t.managerStop()

	t.eventOperationStart()

	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return err
	}

	reader, filename, wait, err := sourceDriver.BackupPipe(t.Context)
	if err != nil {
		return err
	}

	defer func() {
		reader.Close()
		wait() //nolint: errcheck
	}()

	writerList := []io.Writer{}

	for destinationName := range t.TaskConfig.Destinations {
		destinationDriver, err := t.getDestinationDriver(destinationName)
		if err != nil {
			return err
		}

		writer, wait, err := destinationDriver.BackupPipe(t.Context, filename)
		if err != nil {
			return err
		}

		defer func() {
			writer.Close()
			wait() //nolint: errcheck
		}()

		writerList = append(writerList, writer)
	}

	mw := io.MultiWriter(writerList...)

	exitError := make(chan error)

	go func() {
		_, err := io.Copy(mw, reader)

		exitError <- err
	}()

	err = <-exitError

	if err != nil {
		t.eventOperationError(err)
	} else {
		t.eventOperationStop()
	}

	notificationWaitGroup.Wait()

	return err
}

func (t *Task) Backup() error {
	isAvailablePipe, err := t.isAvailableBackupPipe()
	if err != nil {
		return err
	}

	if t.TaskConfig.OperationMode == config.OperationModePipe {
		if !isAvailablePipe {
			return ErrNotSupportedPipeMode
		}

		return t.backupPipe()
	}

	if t.TaskConfig.OperationMode == config.OperationModePreferPipe && isAvailablePipe {
		return t.backupPipe()
	}

	return t.backupManual()
}

func (t *Task) backupWorkDirToDestination(destinationName string) (sctxt.BackupOutput, error) {
	destinationDriver, err := t.getDestinationDriver(destinationName)
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	return destinationDriver.Backup(t.Context)
}

func (t *Task) getDestinationDriver(destinationName string) (interfaces.DriverDestinationInterface, error) {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return nil, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return nil, err
	}

	return destinationDriver, nil
}
