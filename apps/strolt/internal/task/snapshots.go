package task

import (
	"fmt"
	"sort"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

type SnapshotList []interfaces.Snapshot

func (t *Task) GetSnapshotList(destinationName string) (SnapshotList, error) {
	if err := t.managerStart(sctxt.OpTypeSnapshots); err != nil {
		return nil, err
	}
	defer t.managerStop()

	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return nil, fmt.Errorf("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return nil, err
	}

	snapshots, err := destinationDriver.Snapshots()
	if err != nil {
		return nil, err
	}

	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Time.Unix() > snapshots[j].Time.Unix()
	})

	return snapshots, nil
}

func (l SnapshotList) IsAvailable(snapshotID string) bool {
	for _, snapshot := range l {
		if snapshot.ID == snapshotID {
			return true
		}
	}

	return false
}
