package strolt

import (
	"sync"
	"time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

// Manager holds and synchronizes access to all strolt instances.
type Manager struct {
	sync.RWMutex

	Instances map[string]*Instance
}

// ManagerInstanceInit describes the parameters needed to initialize an instance.
type ManagerInstanceInit struct {
	Name     string
	URL      string
	Username string
	Password string
}

// Instance represents a single managed strolt instance.
type Instance struct {
	*sync.RWMutex

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
}

// Config holds the cached configuration of an instance.
type Config struct {
	IsInitialized     bool
	UpdateRequestedAt time.Time
	UpdatedAt         time.Time
	Data              *strolt_models.Config
}

// TaskStatus holds the cached task manager status of an instance.
type TaskStatus struct {
	IsInitialized     bool
	UpdateRequestedAt time.Time
	UpdatedAt         time.Time
	Data              *strolt_models.ManagerStatus
}

// WatchItem tracks the ping and update state of an instance.
type WatchItem struct {
	LatestPingAt            time.Time
	LatestSuccessPingAt     time.Time
	IsPingInProcess         bool
	IsUpdateStatusInProcess bool
}
