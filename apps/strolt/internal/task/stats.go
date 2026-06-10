package task

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// GetStats returns formatted storage statistics for the given destination.
func (t *Task) GetStats(destinationName string) (interfaces.FormattedStats, error) {
	if err := t.managerStart(sctxt.OpTypeStats); err != nil {
		return interfaces.FormattedStats{}, err
	}
	defer t.managerStop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return interfaces.FormattedStats{}, errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return interfaces.FormattedStats{}, fmt.Errorf("get destination driver: %w", err)
	}

	stats, err := destinationDriver.Stats()
	if err != nil {
		return interfaces.FormattedStats{}, fmt.Errorf("get destination stats: %w", err)
	}

	return stats.Convert(), nil
}
