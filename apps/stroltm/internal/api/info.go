package api

import (
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/ldflags"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/strolt"
	"github.com/strolt/strolt/shared/sdk/stroltp"
)

type Info struct {
	Instances     []common.ManagerInfoInstance `json:"instances"`
	UpdatedAt     string                       `json:"updatedAt"`
	LatestVersion string                       `json:"latestVersion"`
	Version       string                       `json:"version"`
}

func getInfoInstances() ([]common.ManagerInfoInstance, int64) {
	stroltInfo := strolt.ManagerGetInfo("")
	stroltpInfo := stroltp.ManagerGetInfo("")

	list := make([]common.ManagerInfoInstance, len(stroltInfo.Instances)+len(stroltpInfo.Instances))

	var updatedAt int64 = 0

	i := 0

	for _, instance := range stroltInfo.Instances {
		list[i] = instance

		_updatedAt := instance.GetUpdatedAt()
		if _updatedAt > updatedAt {
			updatedAt = _updatedAt
		}

		i++
	}

	for _, instance := range stroltpInfo.Instances {
		list[i] = instance

		_updatedAt := instance.GetUpdatedAt()
		if _updatedAt > updatedAt {
			updatedAt = _updatedAt
		}

		i++
	}

	return list, updatedAt
}

// getInfo godoc
// @Id					 getInfo
// @Summary      Get Info
// @Tags         global
// @Security BasicAuth
// @success 200 {object} Info
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	// stroltInstances, stroltUpdatedAt := getStroltInstances()
	// stroltpInstances, stroltpUpdatedAt := getStroltInstancesFromProxy()
	instances, updatedAt := getInfoInstances()

	info := Info{
		Instances:     instances,
		LatestVersion: ldflags.GetVersion(), // TODO: replace this for github release api
		Version:       ldflags.GetVersion(),
	}

	info.UpdatedAt = time.Unix(updatedAt, 0).Format(time.RFC3339)

	apiu.RenderJSON200(w, r, info)
}
