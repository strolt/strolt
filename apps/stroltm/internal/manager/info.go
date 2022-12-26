package manager

import (
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/ldflags"
)

type Info struct {
	Instances     []InfoInstance `json:"instances"`
	UpdatedAt     string         `json:"updatedAt"`
	LatestVersion string         `json:"latestVersion"`
	Version       string         `json:"version"`
}

type InfoInstance struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	LastestOnlineAt string `json:"lastestOnlineAt"`

	StartedAt  string                 `json:"startedAt"`
	IsOnline   bool                   `json:"isOnline"`
	Config     InfoInstanceConfig     `json:"config"`
	TaskStatus InfoInstanceTaskStatus `json:"taskStatus"`
}

type InfoInstanceConfig struct {
	IsInitialized bool   `json:"isInitialized"`
	UpdatedAt     string `json:"updatedAt"`
}

type InfoInstanceTaskStatus struct {
	IsInitialized bool   `json:"isInitialized"`
	UpdatedAt     string `json:"updatedAt"`
}

func GetInfo() Info {
	manager.RLock()
	defer manager.RUnlock()

	var updatedAt int64

	info := Info{
		Instances:     make([]InfoInstance, len(manager.Instances)),
		LatestVersion: ldflags.GetVersion(), // TODO: replace this for github release api
		Version:       ldflags.GetVersion(),
	}

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()

		infoItem := InfoInstance{
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

		infoItem.Config = InfoInstanceConfig{
			IsInitialized: instance.Config.IsInitialized,
			UpdatedAt:     instance.Config.UpdatedAt.Format(time.RFC3339),
		}

		infoItem.TaskStatus = InfoInstanceTaskStatus{
			IsInitialized: instance.TaskStatus.IsInitialized,
			UpdatedAt:     instance.TaskStatus.UpdatedAt.Format(time.RFC3339),
		}

		instance.RUnlock()

		info.Instances[i] = infoItem
		i++
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	return info
}
