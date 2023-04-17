package strolt

import (
	"time"

	"github.com/strolt/strolt/shared/sdk/common"
)

func ManagerGetInfo(version string) common.ManagerInfo {
	manager.RLock()
	defer manager.RUnlock()

	var updatedAt int64

	info := common.ManagerInfo{
		Instances: make([]common.ManagerInfoInstance, len(manager.Instances)),
		Version:   version,
	}

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()

		infoItem := common.ManagerInfoInstance{
			Name:            instance.Name,
			IsOnline:        instance.IsOnline,
			LastestOnlineAt: instance.Watch.LatestSuccessPingAt.Format(time.RFC3339),
		}

		if instance.Info != nil {
			infoItem.Version = instance.Info.Version
			infoItem.StartedAt = instance.Info.StartedAt
		}

		if instance.TaskStatus.UpdateRequestedAt.Unix() > updatedAt {
			updatedAt = instance.TaskStatus.UpdateRequestedAt.Unix()
		}

		if instance.Config.UpdatedAt.Unix() > updatedAt {
			updatedAt = instance.Config.UpdatedAt.Unix()
		}

		infoItem.Config = common.ManagerInfoInstanceConfig{
			IsInitialized: instance.Config.IsInitialized,
			UpdatedAt:     instance.Config.UpdatedAt.Format(time.RFC3339),
		}

		infoItem.TaskStatus = common.ManagerInfoInstanceTaskStatus{
			IsInitialized:     instance.TaskStatus.IsInitialized,
			UpdatedAt:         instance.TaskStatus.UpdatedAt.Format(time.RFC3339),
			UpdateRequestedAt: instance.TaskStatus.UpdateRequestedAt.Format(time.RFC3339),
		}

		if !instance.IsOnline && instance.Watch.LatestSuccessPingAt.Unix() > updatedAt {
			updatedAt = instance.Watch.LatestSuccessPingAt.Unix()
		}

		instance.RUnlock()

		info.Instances[i] = infoItem
		i++
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	return info
}
