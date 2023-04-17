package stroltp

import (
	"time"

	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
)

func getStroltInfo(list []*stroltp_models.ManagerInfoInstance, name string) *stroltp_models.ManagerInfoInstance {
	for _, item := range list {
		if item.Name == name {
			return item
		}
	}

	return nil
}

func ManagerGetInfo(version string) common.ManagerInfo {
	manager.RLock()
	defer manager.RUnlock()

	var updatedAt int64

	info := common.ManagerInfo{
		Instances: []common.ManagerInfoInstance{},
		Version:   version,
	}

	for _, instance := range manager.Instances {
		instance.RLock()

		if instance.Info == nil {
			infoItem := common.ManagerInfoInstance{
				ProxyName:       &instance.Name,
				IsOnline:        false,
				LastestOnlineAt: instance.Watch.LatestSuccessPingAt.Format(time.RFC3339),
			}

			infoItem.Config = common.ManagerInfoInstanceConfig{
				IsInitialized: false,
			}

			infoItem.TaskStatus = common.ManagerInfoInstanceTaskStatus{
				IsInitialized: false,
			}

			info.Instances = append(info.Instances, infoItem)
			instance.RUnlock()

			continue
		}

		for _, stroltInstance := range instance.StroltInstances {
			stroltInfo := getStroltInfo(instance.Info.Instances, stroltInstance.Name)

			latestOnlineAt := instance.Watch.LatestSuccessPingAt.Format(time.RFC3339)

			if stroltInfo != nil {
				latestOnlineAt = stroltInfo.LastestOnlineAt
			}

			infoItem := common.ManagerInfoInstance{
				ProxyName:       &instance.Name,
				IsOnline:        stroltInstance.IsOnline,
				LastestOnlineAt: latestOnlineAt,
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

					infoItem.TaskStatus = common.ManagerInfoInstanceTaskStatus{
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

					infoItem.Config = common.ManagerInfoInstanceConfig{
						IsInitialized: stroltInfo.Config.IsInitialized,
						UpdatedAt:     stroltInfo.Config.UpdatedAt,
					}
				}
			}

			if !instance.IsOnline && instance.Watch.LatestSuccessPingAt.Unix() > updatedAt {
				updatedAt = instance.Watch.LatestSuccessPingAt.Unix()
			}

			info.Instances = append(info.Instances, infoItem)
		}
		instance.RUnlock()
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	return info
}
