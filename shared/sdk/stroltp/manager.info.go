package stroltp

import (
	"time"

	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
)

type ManagerInfo struct {
	Instances []InfoInstance `json:"instances"`
	UpdatedAt string         `json:"updatedAt"`
	Version   string         `json:"version"`
}

type InfoInstance struct {
	ProxyName       string `json:"proxyName,omitempty"`
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
	IsInitialized     bool   `json:"isInitialized"`
	UpdatedAt         string `json:"updatedAt"`
	UpdateRequestedAt string `json:"updateRequestedAt"`
}

func getStroltInfo(list []*stroltp_models.StroltInfoInstance, name string) *stroltp_models.StroltInfoInstance {
	for _, item := range list {
		if item.Name == name {
			return item
		}
	}

	return nil
}

func ManagerGetInfo(version string) ManagerInfo {
	manager.RLock()
	defer manager.RUnlock()

	var updatedAt int64

	info := ManagerInfo{
		Instances: []InfoInstance{},
		Version:   version,
	}

	for _, instance := range manager.Instances {
		instance.RLock()

		if instance.Info == nil {
			infoItem := InfoInstance{
				ProxyName:       instance.Name,
				IsOnline:        false,
				LastestOnlineAt: instance.Watch.LatestSuccessPingAt.Format(time.RFC3339),
			}

			infoItem.Config = InfoInstanceConfig{
				IsInitialized: false,
			}

			infoItem.TaskStatus = InfoInstanceTaskStatus{
				IsInitialized: false,
			}

			info.Instances = append(info.Instances, infoItem)
			instance.RUnlock()

			continue
		}

		for _, stroltInstance := range instance.StroltInstances {
			stroltInfo := getStroltInfo(instance.Info.Instances, stroltInstance.Name)

			infoItem := InfoInstance{
				ProxyName:       instance.Name,
				IsOnline:        stroltInstance.IsOnline,
				LastestOnlineAt: instance.Watch.LatestSuccessPingAt.Format(time.RFC3339),
			}

			if time.Now().Unix() > instance.Watch.LatestSuccessPingAt.Add(time.Second*15).Unix() { //nolint:gomnd
				infoItem.IsOnline = false
			}

			if stroltInfo != nil {
				infoItem.Name = stroltInfo.Name
				infoItem.Version = stroltInfo.Version
				infoItem.StartedAt = stroltInfo.StartedAt

				if stroltInfo.TaskStatus != nil {
					updateRequestedAt, err := time.Parse(time.RFC3339, stroltInfo.TaskStatus.UpdateRequestedAt)
					if err == nil && updateRequestedAt.Unix() > updatedAt {
						updatedAt = updateRequestedAt.Unix()
					}

					infoItem.TaskStatus = InfoInstanceTaskStatus{
						IsInitialized:     stroltInfo.TaskStatus.IsInitialized,
						UpdatedAt:         stroltInfo.TaskStatus.UpdatedAt,
						UpdateRequestedAt: stroltInfo.TaskStatus.UpdateRequestedAt,
					}
				}

				if stroltInfo.Config != nil {
					configUpdatedAt, err := time.Parse(time.RFC3339, stroltInfo.Config.UpdatedAt)

					if err == nil && configUpdatedAt.Unix() > updatedAt {
						updatedAt = configUpdatedAt.Unix()
					}

					infoItem.Config = InfoInstanceConfig{
						IsInitialized: stroltInfo.Config.IsInitialized,
						UpdatedAt:     stroltInfo.Config.UpdatedAt,
					}
				}
			}

			info.Instances = append(info.Instances, infoItem)
		}
		instance.RUnlock()
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	return info
}
