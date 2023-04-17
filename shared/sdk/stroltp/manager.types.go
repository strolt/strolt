package stroltp

import (
	"sync"
	"time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
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
	Info     *stroltp_models.ManagerInfo
	sdk      *SDK
	IsOnline bool

	StroltInstances []*common.ManagerPreparedInstance

	StroltInstancesUpdatedAt int64

	log *logger.Logger
	*sync.RWMutex
}

type WatchItem struct {
	LatestPingAt            time.Time
	LatestSuccessPingAt     time.Time
	IsPingInProcess         bool
	IsUpdateStatusInProcess bool
}
