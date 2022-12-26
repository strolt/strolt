package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t Task) RestoreDestinationToTemp(destinationName string, snapshotName string) error {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return err
	}

	return destinationDriver.Restore(t.Context, snapshotName)
}

func (t Task) RestoreTempToSource() error {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return err
	}

	return sourceDriver.Restore(t.Context)
}

func (t Task) Restore(destinationName string, snapshotName string) error {
	if err := t.managerStart(sctxt.OpTypeRestore); err != nil {
		return err
	}
	defer t.managerStop()

	if err := t.RestoreDestinationToTemp(destinationName, snapshotName); err != nil {
		return err
	}

	return t.RestoreTempToSource()
}
