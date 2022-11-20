package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func (t Task) BackupSourceToWorkDir() error {
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

func (t Task) BackupWorkDirToDestination(destinationName string) (sctxt.BackupOutput, error) {
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
