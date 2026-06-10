package task

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t *Task) isAvailableRestorePipe() (bool, error) {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return false, err
	}

	if !sourceDriver.IsSupportedRestorePipe(t.Context) {
		return false, nil
	}

	for destinationName := range t.TaskConfig.Destinations {
		destinationDriver, err := t.getDestinationDriver(destinationName)
		if err != nil {
			return false, err
		}

		if !destinationDriver.IsSupportedRestorePipe(t.Context) {
			return false, nil
		}
	}

	return true, nil
}

// RestoreDestinationToTemp restores a snapshot from the destination into the work directory.
func (t *Task) RestoreDestinationToTemp(destinationName string, snapshotName string) error {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return fmt.Errorf("get destination driver: %w", err)
	}

	if err := destinationDriver.Restore(t.Context, snapshotName); err != nil {
		return fmt.Errorf("destination restore: %w", err)
	}

	return nil
}

// RestoreTempToSource restores data from the work directory into the source.
func (t *Task) RestoreTempToSource() error {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return fmt.Errorf("get source driver: %w", err)
	}

	if err := sourceDriver.Restore(t.Context); err != nil {
		return fmt.Errorf("source restore: %w", err)
	}

	return nil
}

// restorePipe streams a snapshot from the destination directly into the
// source. The exit status of both processes must fail the restore: a crashed
// reader or writer with a closed stream would otherwise look like a
// successful restore of truncated data.
func (t *Task) restorePipe(destinationName string, snapshotName string) error {
	destinationDriver, err := t.getDestinationDriver(destinationName)
	if err != nil {
		return err
	}

	reader, filename, destinationWait, err := destinationDriver.RestorePipe(t.Context, snapshotName)
	if err != nil {
		return fmt.Errorf("destination restore pipe: %w", err)
	}

	waitDestination := sync.OnceValue(destinationWait)

	defer func() {
		_ = reader.Close()
		_ = waitDestination()
	}()

	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return err
	}

	writer, sourceWait, err := sourceDriver.RestorePipe(t.Context, filename)
	if err != nil {
		return fmt.Errorf("source restore pipe: %w", err)
	}

	waitSource := sync.OnceValue(sourceWait)

	defer func() {
		_ = writer.Close()
		_ = waitSource()
	}()

	_, copyErr := io.Copy(writer, reader)

	errs := []error{copyErr}

	if err := waitDestination(); err != nil {
		errs = append(errs, fmt.Errorf("destination process: %w", err))
	}

	// Close the writer so the source process sees EOF before its exit
	// status is collected.
	errs = append(errs, writer.Close())

	if err := waitSource(); err != nil {
		errs = append(errs, fmt.Errorf("source process: %w", err))
	}

	return errors.Join(errs...)
}

func (t *Task) restoreCopy(destinationName string, snapshotName string) error {
	if err := t.RestoreDestinationToTemp(destinationName, snapshotName); err != nil {
		return err
	}

	return t.RestoreTempToSource()
}

// Restore runs the restore operation in pipe or copy mode depending on the task config.
func (t *Task) Restore(destinationName string, snapshotName string) error {
	isAvailablePipe, err := t.isAvailableRestorePipe()
	if err != nil {
		return err
	}

	if err := t.managerStart(sctxt.OpTypeRestore); err != nil {
		return err
	}
	defer t.managerStop()

	if t.TaskConfig.OperationMode == config.OperationModePipe {
		if !isAvailablePipe {
			return ErrNotSupportedPipeMode
		}

		return t.restorePipe(destinationName, snapshotName)
	}

	if t.TaskConfig.OperationMode == config.OperationModePreferPipe && isAvailablePipe {
		return t.restorePipe(destinationName, snapshotName)
	}

	return t.restoreCopy(destinationName, snapshotName)
}
