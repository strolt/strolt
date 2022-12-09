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
	Strolt map[string]*Strolt
}

type Strolt struct {
	InstanceName string
	URL          string
	Username     string
	Password     string

	Watch          *WatchItem
	Info           *InfoItem
	ConfigLoadedAt time.Time
	IsOnline       bool
	sdk            *strolt.Sdk

	Config *models.APIConfig

	log *logger.Logger

	*sync.Mutex
}

type WatchItem struct {
	LatestPingAt        time.Time
	LatestSuccessPingAt time.Time
	IsPingInProcess     bool
}

type InfoItem struct {
	Version string
}

var (
	manager = &Manager{
		Strolt: map[string]*Strolt{},
	}
)

func New() *Manager {
	configInstances := config.Get().Strolt.Instances

	for instanceName, instance := range configInstances {
		manager.Strolt[instanceName] = &Strolt{
			Watch:        &WatchItem{},
			Info:         &InfoItem{},
			InstanceName: instanceName,
			URL:          instance.URL,
			Username:     instance.Username,
			Password:     instance.Password, //pragma: allowlist secret
			sdk:          strolt.New(instance.URL, instance.Username, instance.Password),
			log:          logger.New().WithField("instanceName", instanceName),
			Mutex:        &sync.Mutex{},
		}
	}

	logger.New().Infof("initialized %d strolt instances", len(configInstances))

	return manager
}

func GetStroltInstances() []Strolt {
	list := make([]Strolt, len(manager.Strolt))
	i := 0

	for _, instance := range manager.Strolt {
		list[i] = *instance
		i++
	}

	return list
}

func GetStroltInstanceByInstanceName(instanceName string) (Strolt, bool) {
	instance, ok := manager.Strolt[instanceName]
	if !ok {
		return Strolt{}, ok
	}

	return *instance, ok
}
