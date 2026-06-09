// Package common contains shared manager types used by strolt SDKs.
package common

import (
	"time"
)

// ManagerInfo describes the aggregated state of all managed instances.
type ManagerInfo struct {
	Instances []ManagerInfoInstance `json:"instances"`
	UpdatedAt string                `json:"updatedAt"`
	Version   string                `json:"version"`
} //	@name	ManagerInfo

// ManagerInfoInstance describes the state of a single managed instance.
type ManagerInfoInstance struct {
	ProxyName       *string `json:"proxyName,omitempty"`
	Name            string  `json:"name"`
	Version         string  `json:"version"`
	LastestOnlineAt string  `json:"lastestOnlineAt"`

	StartedAt  string                        `json:"startedAt"`
	IsOnline   bool                          `json:"isOnline"`
	Config     ManagerInfoInstanceConfig     `json:"config"`
	TaskStatus ManagerInfoInstanceTaskStatus `json:"taskStatus"`
} //	@name	ManagerInfoInstance

// ManagerInfoInstanceConfig describes the configuration state of an instance.
type ManagerInfoInstanceConfig struct {
	IsInitialized bool   `json:"isInitialized"`
	UpdatedAt     string `json:"updatedAt"`
} //	@name	ManagerInfoInstanceConfig

// ManagerInfoInstanceTaskStatus describes the task status state of an instance.
type ManagerInfoInstanceTaskStatus struct {
	IsInitialized     bool   `json:"isInitialized"`
	UpdatedAt         string `json:"updatedAt"`
	UpdateRequestedAt string `json:"updateRequestedAt"`
} //	@name	ManagerInfoInstanceTaskStatus

// GetUpdatedAt returns the latest update timestamp of the instance as a Unix time.
func (instance *ManagerInfoInstance) GetUpdatedAt() int64 {
	var updatedAt int64

	taskStatusUpdatedRequestedAt, err := time.Parse(time.RFC3339, instance.TaskStatus.UpdateRequestedAt)
	if err == nil {
		if taskStatusUpdatedRequestedAt.Unix() > updatedAt {
			updatedAt = taskStatusUpdatedRequestedAt.Unix()
		}
	}

	taskStatusUpdatedAt, err := time.Parse(time.RFC3339, instance.Config.UpdatedAt)
	if err == nil {
		if taskStatusUpdatedAt.Unix() > updatedAt {
			updatedAt = taskStatusUpdatedAt.Unix()
		}
	}

	return updatedAt
}

// type ManagerPreparedInstance struct {
// 	ProxyName  string                                  `json:"proxyName"`
// 	Name       string                                  `json:"name"`
// 	Config     *ManagerPreparedInstanceConfig         `json:"config"`
// 	TaskStatus *stroltp_models.StroltTaskManagerStatus `json:"taskStatus"`
// 	IsOnline   bool                                    `json:"isOnline"`
// }// @name ManagerPreparedInstance

// type ManagerPreparedInstanceConfig struct {
// 	TimeZone            string                   `json:"timezone"`
// 	DisableWatchChanges bool                     `json:"disableWatchChanges"`
// 	Tags                []string                 `json:"tags"`
// 	Services            map[string]ManagerPreparedInstanceConfigService `json:"services"`
// }// @name ManagerPreparedInstanceConfig

// type ManagerPreparedInstanceConfigService map[string]ManagerPreparedInstanceConfigServiceTask // @name ManagerPreparedInstanceConfigService

// type ManagerPreparedInstanceConfigServiceTask struct {
// 	Source        ManagerPreparedInstanceConfigServiceTaskSource                 `json:"source"`
// 	Destinations  map[string]ManagerPreparedInstanceConfigServiceTaskDestination `json:"destinations"`
// 	Notifications []ManagerPreparedInstanceConfigServiceTaskNotification         `json:"notifications"`
// 	Schedule      ManagerPreparedInstanceConfigServiceTaskSchedule               `json:"schedule"`
// 	Tags          []string                                `json:"tags"`
// } // @name ManagerPreparedInstanceConfigServiceTask

// type ManagerPreparedInstanceConfigServiceTaskSource struct {
// 	Driver string `json:"driver"`
// } // @name ManagerPreparedInstanceConfigServiceTaskSource

// type ManagerPreparedInstanceConfigServiceTaskSchedule struct {
// 	Backup string `json:"backup"`
// 	Prune  string `json:"prune"`
// } // @name ManagerPreparedInstanceConfigServiceTaskSchedule

// type ManagerPreparedInstanceConfigServiceTaskDestination struct {
// 	Driver string `json:"driver"`
// } // @name ManagerPreparedInstanceConfigServiceTaskDestination

// type ManagerPreparedInstanceConfigServiceTaskNotification struct {
// 	Driver string            `json:"driver"`
// 	Name   string            `json:"name"`
// 	Events []sctxt.EventType `json:"events" enums:"OPERATION_START,OPERATION_STOP,OPERATION_ERROR,SOURCE_START,SOURCE_STOP,SOURCE_ERROR,DESTINATION_START,DESTINATION_STOP,DESTINATION_ERROR"`
// } // @name ManagerPreparedInstanceConfigServiceTaskNotification

// type ManagerPreparedInstanceTaskManagerStatus = strolt_models.ManagerStatus
