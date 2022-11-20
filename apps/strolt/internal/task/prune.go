package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (t Task) DestinationPrune(destinationName string, isDryRun bool) ([]interfaces.Snapshot, error) {
	operation := ControllerOperation{
		ServiceName:     t.ServiceName,
		TaskName:        t.TaskName,
		DestinationName: destinationName,
		Operation:       COPruneDestination,
	}

	if err := operation.Start(); err != nil {
		return []interfaces.Snapshot{}, err
	}
	defer operation.Stop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return []interfaces.Snapshot{}, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return []interfaces.Snapshot{}, err
	}

	return destinationDriver.Prune(t.Context, isDryRun)
}
