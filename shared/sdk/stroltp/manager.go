package stroltp

import (
	"context"
	"sync"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/common"
)

var (
	manager = &Manager{
		Instances: map[string]*Instance{},
	}
)

func ManagerInit(ctx context.Context, cancel func(), instances []ManagerInstanceInit) {
	for _, instance := range instances {
		manager.Instances[instance.Name] = &Instance{
			Name:                     instance.Name,
			URL:                      instance.URL,
			Username:                 instance.Username,
			Password:                 instance.Password, //pragma: allowlist secret
			Watch:                    WatchItem{},
			sdk:                      New(instance.URL, instance.Username, instance.Password),
			log:                      logger.New().WithField("proxyInstanceName", instance.Name),
			RWMutex:                  &sync.RWMutex{},
			StroltInstances:          []*common.ManagerPreparedInstance{},
			StroltInstancesUpdatedAt: 0,
		}
	}

	logger.New().Infof("initialized %d stroltp instance(s)", len(instances))

	manager.Watch(ctx, cancel)
}

func ManagerGetPreparedInstances() []common.ManagerPreparedInstance {
	manager.RLock()
	defer manager.RUnlock()

	countInstances := 0

	for _, instance := range manager.Instances {
		instance.RLock()

		for range instance.StroltInstances {
			countInstances++
		}

		instance.RUnlock()
	}

	list := make([]common.ManagerPreparedInstance, countInstances)

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()

		for _, stroltInstance := range instance.StroltInstances {
			list[i] = common.ManagerPreparedInstance{
				ProxyName:  &instance.Name,
				Name:       stroltInstance.Name,
				Config:     stroltInstance.Config,
				TaskStatus: stroltInstance.TaskStatus,
				IsOnline:   stroltInstance.IsOnline,
			}

			i++
		}

		instance.RUnlock()
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
