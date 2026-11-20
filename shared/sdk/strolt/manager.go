package strolt

import (
	"context"
	"sync"
	time "time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/common"
)

var manager = &Manager{
	RWMutex:   sync.RWMutex{},
	Instances: map[string]*Instance{},
}

// ManagerInit initializes the manager with the given instances and starts watching them.
func ManagerInit(ctx context.Context, cancel func(), instances []ManagerInstanceInit) {
	for _, instance := range instances {
		manager.Instances[instance.Name] = &Instance{
			Info: nil, IsOnline: false,
			Name:       instance.Name,
			URL:        instance.URL,
			Username:   instance.Username,
			Password:   instance.Password, // pragma: allowlist secret
			Watch:      WatchItem{LatestPingAt: time.Time{}, LatestSuccessPingAt: time.Time{}, IsPingInProcess: false, IsUpdateStatusInProcess: false},
			sdk:        New(instance.URL, instance.Username, instance.Password),
			log:        logger.New().WithField("instanceName", instance.Name),
			RWMutex:    &sync.RWMutex{},
			Config:     Config{IsInitialized: false, UpdateRequestedAt: time.Time{}, UpdatedAt: time.Time{}, Data: nil},
			TaskStatus: TaskStatus{IsInitialized: false, UpdateRequestedAt: time.Time{}, UpdatedAt: time.Time{}, Data: nil},
		}
	}

	logger.New().Infof("initialized %d strolt instance(s)", len(instances))

	manager.Watch(ctx, cancel)
}

// ManagerGetPreparedInstances returns a snapshot of all managed instances.
func ManagerGetPreparedInstances() []common.ManagerPreparedInstance {
	manager.RLock()
	defer manager.RUnlock()

	list := make([]common.ManagerPreparedInstance, len(manager.Instances))

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()
		list[i] = common.ManagerPreparedInstance{
			ProxyName:  nil,
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

// ManagerGetSDKByInstanceName returns the SDK client for the named instance.
func ManagerGetSDKByInstanceName(instanceName string) (*SDK, bool) {
	instance, ok := manager.Instances[instanceName]
	if !ok {
		return nil, ok
	}

	return instance.sdk, ok
}
