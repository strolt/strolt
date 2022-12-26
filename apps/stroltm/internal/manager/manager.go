package manager

import (
	"sync"
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/config"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/models"
	"github.com/strolt/strolt/shared/logger"
)

type Manager struct {
	Instances map[string]*Instance
	sync.RWMutex
}

var (
	manager = &Manager{
		Instances: map[string]*Instance{},
	}
)

type Instance struct {
	Name     string
	URL      string
	Username string
	Password string

	Watch    WatchItem
	Info     *models.APIGetInfoResponse
	Online   bool
	sdk      *strolt.SDK
	IsOnline bool

	TaskStatus TaskStatus

	Config Config

	log *logger.Logger

	*sync.RWMutex
}

type Config struct {
	IsInitialized     bool
	UpdateRequestedAt time.Time
	UpdatedAt         time.Time
	Data              *models.APIConfig
}

type TaskStatus struct {
	IsInitialized     bool
	UpdateRequestedAt time.Time
	UpdatedAt         time.Time
	Data              *models.TaskManagerStatus
}

type WatchItem struct {
	LatestPingAt                time.Time
	LatestSuccessPingAt         time.Time
	IsPingInProcess             bool
	LatestSuccessUpdateStatusAt time.Time
	IsUpdateStatusInProcess     bool
}

func Init() *Manager {
	configInstances := config.Get().Strolt.Instances

	for instanceName, instance := range configInstances {
		manager.Instances[instanceName] = &Instance{
			Name:       instanceName,
			URL:        instance.URL,
			Username:   instance.Username,
			Password:   instance.Password, //pragma: allowlist secret
			Watch:      WatchItem{},
			sdk:        strolt.New(instance.URL, instance.Username, instance.Password),
			log:        logger.New().WithField("instanceName", instanceName),
			RWMutex:    &sync.RWMutex{},
			Config:     Config{},
			TaskStatus: TaskStatus{},
		}
	}

	logger.New().Infof("initialized %d strolt instances", len(configInstances))

	return manager
}

func GetSDKByInstanceName(instanceName string) (*strolt.SDK, bool) {
	instance, ok := manager.Instances[instanceName]
	if !ok {
		return nil, ok
	}

	return instance.sdk, ok
}

type PreparedInstance struct {
	Name       string                    `json:"name"`
	Config     *models.APIConfig         `json:"config"`
	TaskStatus *models.TaskManagerStatus `json:"taskStatus"`
	IsOnline   bool                      `json:"isOnline"`
}

func GetPreparedInstances() []PreparedInstance {
	manager.RLock()
	list := make([]PreparedInstance, len(manager.Instances))

	i := 0

	for _, instance := range manager.Instances {
		instance.RLock()
		list[i] = PreparedInstance{
			Name:       instance.Name,
			Config:     instance.Config.Data,
			TaskStatus: instance.TaskStatus.Data,
			IsOnline:   instance.IsOnline,
		}
		instance.RUnlock()

		i++
	}

	manager.RUnlock()

	return list
}
