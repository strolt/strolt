package destinations

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"

	"github.com/go-chi/chi/v5"
)

type getSnapshotsResult struct {
	Data        task.SnapshotList `json:"data"`
	LastUpdated time.Time         `json:"lastUpdated"`
}

type getSnapshotsCache struct {
	sync.Mutex
	List []getSnapshotsCacheItem
}

type getSnapshotsCacheItem struct {
	ServiceName     string
	TaskName        string
	DestinationName string
	SnapshotList    task.SnapshotList
	LastUpdated     time.Time
}

var snapshotsCache = getSnapshotsCache{}

func getSnapshotsFromCache(serviceName string, taskName string, destinationName string) (getSnapshotsCacheItem, error) {
	if snapshotsCache.List == nil {
		return getSnapshotsCacheItem{}, fmt.Errorf("cache is empty")
	}

	for _, item := range snapshotsCache.List {
		if item.ServiceName == serviceName &&
			item.TaskName == taskName &&
			item.DestinationName == destinationName {
			return item, nil
		}
	}

	return getSnapshotsCacheItem{}, nil
}

func addSnapshotsToCache(serviceName string, taskName string, destinationName string, snapshotList task.SnapshotList) getSnapshotsCacheItem {
	snapshotsCache.Lock()

	snapshotsItem := getSnapshotsCacheItem{
		ServiceName:     serviceName,
		TaskName:        taskName,
		DestinationName: destinationName,
		SnapshotList:    snapshotList,
		LastUpdated:     time.Now(),
	}

	isExistsInCache := false
	existsIndex := 0

	for i, item := range snapshotsCache.List {
		if item.ServiceName == snapshotsItem.ServiceName &&
			item.TaskName == snapshotsItem.TaskName &&
			item.DestinationName == snapshotsItem.DestinationName {
			isExistsInCache = true
			existsIndex = i

			break
		}
	}

	if isExistsInCache {
		snapshotsCache.List[existsIndex] = snapshotsItem
	} else {
		snapshotsCache.List = append(snapshotsCache.List, snapshotsItem)
	}

	snapshotsCache.Unlock()

	return snapshotsItem
}

// hGetSnapshots godoc
// @Id					 getSnapshots
// @Summary      Get snapshots
// @Tags         services
// @Accept       json
// @Produce      json
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getSnapshotsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots [get].
func getSnapshots(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	taskOperation := task.ControllerOperation{
		ServiceName:     serviceName,
		TaskName:        taskName,
		DestinationName: destinationName,
	}

	if taskOperation.IsWorking() {
		cacheItem, err := getSnapshotsFromCache(serviceName, taskName, destinationName)
		if err == nil {
			apiu.RenderJSON200(w, r, getSnapshotsResult{
				Data:        cacheItem.SnapshotList,
				LastUpdated: cacheItem.LastUpdated,
			})

			return
		}

		apiu.RenderJSON500(w, r, apiu.ResultError{Error: apiu.ErrTaskAlreadyWorking.Error()})

		return
	}

	t, err := task.New(serviceName, taskName, sctxt.TApi, sctxt.OpTypeBackup)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}
	defer t.Close()

	snapshotList, err := t.GetSnapshotList(destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	cacheItem := addSnapshotsToCache(serviceName, taskName, destinationName, snapshotList)

	apiu.RenderJSON200(w, r, getSnapshotsResult{
		Data:        cacheItem.SnapshotList,
		LastUpdated: cacheItem.LastUpdated,
	})
}
