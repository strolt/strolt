package strolt

import (
	"context"
	"sync"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

var (
	manager = &Manager{
		Instances: map[string]*Instance{},
	}
)

func ManagerInit(ctx context.Context, cancel func(), instances []ManagerInstanceInit) {
	for _, instance := range instances {
		manager.Instances[instance.Name] = &Instance{
			Name:       instance.Name,
			URL:        instance.URL,
			Username:   instance.Username,
			Password:   instance.Password, //pragma: allowlist secret
			Watch:      WatchItem{},
			sdk:        New(instance.URL, instance.Username, instance.Password),
			log:        logger.New().WithField("instanceName", instance.Name),
			RWMutex:    &sync.RWMutex{},
			Config:     Config{},
			TaskStatus: TaskStatus{},
		}
	}

	logger.New().Infof("initialized %d strolt instance(s)", len(instances))

	manager.Watch(ctx, cancel)
}

type APIConfig = strolt_models.Config
type TaskManagerStatus = strolt_models.ManagerStatus

type ManagerPreparedInstance struct {
	Name       string             `json:"name"`
	Config     *APIConfig         `json:"config"`
	TaskStatus *TaskManagerStatus `json:"taskStatus"`
	IsOnline   bool               `json:"isOnline"`
}

func ManagerGetPreparedInstances() []ManagerPreparedInstance {
	manager.RLock()
	defer manager.RUnlock()

	list := make([]ManagerPreparedInstance, len(manager.Instances))

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()
		list[i] = ManagerPreparedInstance{
			Name:       instance.Name,
			Config:     instance.Config.Data,
			TaskStatus: instance.TaskStatus.Data,
			IsOnline:   instance.IsOnline,
		}
		instance.RUnlock()

		i++
	}

	return list
}

func ManagerGetSDKByInstanceName(instanceName string) (*SDK, bool) {
	instance, ok := manager.Instances[instanceName]
	if !ok {
		return nil, ok
	}

	return instance.sdk, ok
}
