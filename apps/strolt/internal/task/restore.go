package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
)

func (t Task) RestoreDestinationToTemp(destinationName string, snapshotName string) error {
	operation := ControllerOperation{
		ServiceName:     t.ServiceName,
		TaskName:        t.TaskName,
		DestinationName: destinationName,
		Operation:       CORestoreDestination,
	}

	if err := operation.Start(); err != nil {
		return err
	}
	defer operation.Stop()

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
	operation := ControllerOperation{
		ServiceName: t.ServiceName,
		TaskName:    t.TaskName,
		Operation:   CORestoreSource,
	}

	if err := operation.Start(); err != nil {
		return err
	}
	defer operation.Stop()

	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return err
	}

	return sourceDriver.Restore(t.Context)
}

func (t Task) Restore(destinationName string, snapshotName string) error {
	if err := t.RestoreDestinationToTemp(destinationName, snapshotName); err != nil {
		return err
	}

	return t.RestoreTempToSource()
}
