package task

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// Unlock removes locks left in the destination repository. By default only
// stale locks are removed; isRemoveAll also removes the locks of operations
// that are still running elsewhere.
func (t *Task) Unlock(destinationName string, isRemoveAll bool) error {
	if err := t.managerStart(sctxt.OpTypeUnlock); err != nil {
		return err
	}
	defer t.managerStop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return fmt.Errorf("get destination driver: %w", err)
	}

	if err := destinationDriver.Unlock(isRemoveAll); err != nil {
		return fmt.Errorf("destination unlock: %w", err)
	}

	return nil
}
