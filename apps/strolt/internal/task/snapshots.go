package task

import (
	"errors"
	"fmt"
	"sort"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// SnapshotList is a list of snapshots stored in a destination.
type SnapshotList []interfaces.Snapshot

// GetSnapshotList returns the snapshots of the given destination sorted by time descending.
func (t *Task) GetSnapshotList(destinationName string) (SnapshotList, error) {
	if err := t.managerStart(sctxt.OpTypeSnapshots); err != nil {
		return nil, err
	}
	defer t.managerStop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return nil, errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return nil, fmt.Errorf("get destination driver: %w", err)
	}

	snapshots, err := destinationDriver.Snapshots()
	if err != nil {
		return nil, fmt.Errorf("get destination snapshots: %w", err)
	}

	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Time.Unix() > snapshots[j].Time.Unix()
	})

	return snapshots, nil
}

// IsAvailable reports whether a snapshot with the given ID exists in the list.
func (l SnapshotList) IsAvailable(snapshotID string) bool {
	for _, snapshot := range l {
		if snapshot.ID == snapshotID {
			return true
		}
	}

	return false
}
