package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (t Task) GetStats(destinationName string) (interfaces.FormattedStats, error) {
	operation := ControllerOperation{
		ServiceName:     t.ServiceName,
		TaskName:        t.TaskName,
		DestinationName: destinationName,
		Operation:       COFetchStats,
	}

	if err := operation.Start(); err != nil {
		return interfaces.FormattedStats{}, err
	}
	defer operation.Stop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return interfaces.FormattedStats{}, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return interfaces.FormattedStats{}, err
	}

	stats, err := destinationDriver.Stats()
	if err != nil {
		return interfaces.FormattedStats{}, err
	}

	return stats.Convert(), nil
}
