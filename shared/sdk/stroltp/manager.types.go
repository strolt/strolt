package stroltp

import (
	"sync"
	"time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
)

// Manager keeps track of all configured strolt proxy instances.
type Manager struct {
	sync.RWMutex

	Instances map[string]*Instance
}

// ManagerInstanceInit holds the parameters required to register a proxy instance in the manager.
type ManagerInstanceInit struct {
	Name     string
	URL      string
	Username string
	Password string
}

// Instance represents a single strolt proxy instance tracked by the manager.
type Instance struct {
	*sync.RWMutex

	Name     string
	URL      string
	Username string
	Password string

	Watch    WatchItem
	Info     *stroltp_models.ManagerInfo
	sdk      *SDK
	IsOnline bool

	StroltInstances []*common.ManagerPreparedInstance

	StroltInstancesUpdatedAt int64

	log *logger.Logger
}

// WatchItem holds the ping state of a proxy instance.
type WatchItem struct {
	LatestPingAt            time.Time
	LatestSuccessPingAt     time.Time
	IsPingInProcess         bool
	IsUpdateStatusInProcess bool
}
