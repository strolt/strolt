package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (t Task) GetStats(destinationName string) (interfaces.Stats, error) {
	operation := ControllerOperation{
		ServiceName:     t.ServiceName,
		TaskName:        t.TaskName,
		DestinationName: destinationName,
		Operation:       COFetchStats,
	}

	if err := operation.Start(); err != nil {
		return interfaces.Stats{}, err
	}
	defer operation.Stop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return interfaces.Stats{}, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return interfaces.Stats{}, err
	}

	stats, err := destinationDriver.Stats()
	if err != nil {
		return interfaces.Stats{}, err
	}

	return stats, nil
}
