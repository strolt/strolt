package strolt

import (
	"sync"
	"time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

type Manager struct {
	Instances map[string]*Instance
	sync.RWMutex
}

type ManagerInstanceInit struct {
	Name     string
	URL      string
	Username string
	Password string
}

type Instance struct {
	Name     string
	URL      string
	Username string
	Password string

	Watch    WatchItem
	Info     *strolt_models.APIGetInfoResponse
	sdk      *SDK
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
	Data              *strolt_models.Config
}

type TaskStatus struct {
	IsInitialized     bool
	UpdateRequestedAt time.Time
	UpdatedAt         time.Time
	Data              *strolt_models.ManagerStatus
}

type WatchItem struct {
	LatestPingAt            time.Time
	LatestSuccessPingAt     time.Time
	IsPingInProcess         bool
	IsUpdateStatusInProcess bool
}
