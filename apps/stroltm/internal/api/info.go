package api

import (
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/ldflags"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/sdk/strolt"
	"github.com/strolt/strolt/shared/sdk/stroltp"
)

type Info struct {
	Instances     []InfoInstance `json:"instances"`
	UpdatedAt     string         `json:"updatedAt"`
	LatestVersion string         `json:"latestVersion"`
	Version       string         `json:"version"`
}

type InfoInstance struct {
	ProxyName       *string `json:"proxyName,omitempty"`
	Name            string  `json:"name"`
	Version         string  `json:"version"`
	LastestOnlineAt string  `json:"lastestOnlineAt"`

	StartedAt  string                 `json:"startedAt"`
	IsOnline   bool                   `json:"isOnline"`
	Config     InfoInstanceConfig     `json:"config"`
	TaskStatus InfoInstanceTaskStatus `json:"taskStatus"`
}

type InfoInstanceConfig struct {
	IsInitialized bool   `json:"isInitialized"`
	UpdatedAt     string `json:"updatedAt"`
}

type InfoInstanceTaskStatus struct {
	IsInitialized bool   `json:"isInitialized"`
	UpdatedAt     string `json:"updatedAt"`
}

func getStroltInstances() ([]InfoInstance, int64) {
	stroltInfo := strolt.ManagerGetInfo("")

	var updatedAt int64

	instances := make([]InfoInstance, len(stroltInfo.Instances))

	for i, instance := range stroltInfo.Instances {
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

		instances[i] = InfoInstance{
			Name:            instance.Name,
			Version:         instance.Version,
			LastestOnlineAt: instance.LastestOnlineAt,

			StartedAt: instance.StartedAt,
			IsOnline:  instance.IsOnline,

			Config: InfoInstanceConfig{
				IsInitialized: instance.Config.IsInitialized,
				UpdatedAt:     instance.Config.UpdatedAt,
			},

			TaskStatus: InfoInstanceTaskStatus{
				IsInitialized: instance.TaskStatus.IsInitialized,
				UpdatedAt:     instance.TaskStatus.UpdatedAt,
			},
		}
	}

	return instances, updatedAt
}

func getStroltInstancesFromProxy() ([]InfoInstance, int64) {
	stroltInfo := stroltp.ManagerGetInfo("")

	var updatedAt int64

	instances := make([]InfoInstance, len(stroltInfo.Instances))

	for i, instance := range stroltInfo.Instances {
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

		proxyName := instance.ProxyName

		instances[i] = InfoInstance{
			ProxyName:       &proxyName,
			Name:            instance.Name,
			Version:         instance.Version,
			LastestOnlineAt: instance.LastestOnlineAt,

			StartedAt: instance.StartedAt,
			IsOnline:  instance.IsOnline,

			Config: InfoInstanceConfig{
				IsInitialized: instance.Config.IsInitialized,
				UpdatedAt:     instance.Config.UpdatedAt,
			},

			TaskStatus: InfoInstanceTaskStatus{
				IsInitialized: instance.TaskStatus.IsInitialized,
				UpdatedAt:     instance.TaskStatus.UpdatedAt,
			},
		}
	}

	return instances, updatedAt
}

// getInfo godoc
// @Id					 getInfo
// @Summary      Get Info
// @Tags         global
// @Security BasicAuth
// @success 200 {object} Info
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	stroltInstances, stroltUpdatedAt := getStroltInstances()
	stroltpInstances, stroltpUpdatedAt := getStroltInstancesFromProxy()

	var updatedAt int64

	info := Info{
		Instances:     make([]InfoInstance, len(stroltInstances)+len(stroltpInstances)),
		LatestVersion: ldflags.GetVersion(), // TODO: replace this for github release api
		Version:       ldflags.GetVersion(),
	}

	if stroltUpdatedAt > updatedAt {
		updatedAt = stroltUpdatedAt
	}

	if stroltpUpdatedAt > updatedAt {
		updatedAt = stroltpUpdatedAt
	}

	i := 0

	for _, instance := range stroltInstances {
		info.Instances[i] = instance
		i++
	}

	for _, instance := range stroltpInstances {
		info.Instances[i] = instance
		i++
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	apiu.RenderJSON200(w, r, info)
}
