package task

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t *Task) forget(destinationName string, snapshotID string) error {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return fmt.Errorf("get destination driver: %w", err)
	}

	if err := destinationDriver.Forget(snapshotID); err != nil {
		return fmt.Errorf("destination forget: %w", err)
	}

	return nil
}

// Forget removes a single snapshot from the given destination, ignoring the
// retention policy the destination is configured with.
func (t *Task) Forget(destinationName string, snapshotID string) error {
	if err := t.managerStart(sctxt.OpTypeForget); err != nil {
		return err
	}
	defer t.managerStop()

	t.eventOperationStart()
	t.eventDestinationStart(destinationName)

	err := t.forget(destinationName, snapshotID)
	if err != nil {
		t.eventDestinationError(destinationName, err)
		t.eventOperationError(err)
	} else {
		t.eventDestinationStop(destinationName, sctxt.BackupOutput{})
		t.eventOperationStop()
	}

	t.notifyWaitGroup().Wait()

	return err
}
