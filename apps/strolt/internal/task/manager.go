package task

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

type manager struct {
	Tasks         map[string]ManagerTaskItem
	LastChangedAt time.Time
	*sync.RWMutex
}

type ManagerTaskItem struct {
	ServiceName string              `json:"serviceName"`
	TaskName    string              `json:"taskName"`
	Opeation    sctxt.OperationType `json:"operation"`
	StartedAt   time.Time           `json:"startedAt"`
	LastEndedAt time.Time           `json:"lastEndedAt"`
	TriggerType sctxt.TriggerType   `json:"trigger"`
	IsRunning   bool                `json:"isRunning"`
} // @name ManagerTaskItem

type ManagerStatus struct {
	Tasks         []ManagerTaskItem `json:"tasks"`
	LastChangedAt string            `json:"lastChangedAt"`
} // @name ManagerStatus

var managerVar = manager{
	Tasks:         map[string]ManagerTaskItem{},
	LastChangedAt: time.Now(),
	RWMutex:       &sync.RWMutex{},
}

func (t *Task) managerStart(operation sctxt.OperationType) error {
	taskKey := t.mangerGetKeyForTask()

	if t.IsRunning() {
		return t.managerCreateErrorIsRunning()
	}

	managerVar.Lock()

	item := ManagerTaskItem{
		ServiceName: t.Context.ServiceName,
		TaskName:    t.Context.TaskName,
		IsRunning:   true,
		StartedAt:   time.Now(),
		LastEndedAt: time.Now(),
		Opeation:    operation,
		TriggerType: t.Trigger,
	}

	taskItem, ok := managerVar.Tasks[taskKey]
	if ok {
		item.LastEndedAt = taskItem.LastEndedAt
	}

	managerVar.Tasks[taskKey] = item

	managerVar.LastChangedAt = time.Now()

	managerVar.Unlock()

	return nil
}

func (t *Task) managerStop() {
	taskKey := t.mangerGetKeyForTask()

	managerVar.Lock()

	item := ManagerTaskItem{
		ServiceName: t.Context.ServiceName,
		TaskName:    t.Context.TaskName,
		IsRunning:   false,
		LastEndedAt: time.Now(),
	}

	taskItem, ok := managerVar.Tasks[taskKey]
	if ok {
		item.StartedAt = taskItem.StartedAt
		item.Opeation = taskItem.Opeation
	}

	managerVar.Tasks[taskKey] = item

	managerVar.LastChangedAt = time.Now()

	managerVar.Unlock()
}

func (t *Task) IsRunning() bool {
	managerVar.RLock()
	defer managerVar.RUnlock()

	taskKey := t.mangerGetKeyForTask()

	taskItem, ok := managerVar.Tasks[taskKey]
	if !ok {
		return false
	}

	return taskItem.IsRunning
}

func GetLastChangedManager() time.Time {
	managerVar.RLock()
	defer managerVar.RUnlock()

	return managerVar.LastChangedAt
}

func GetManagerStatus() ManagerStatus {
	status := ManagerStatus{}
	list := []ManagerTaskItem{}

	managerVar.RLock()
	status.LastChangedAt = managerVar.LastChangedAt.Format(time.RFC3339)

	for _, taskItem := range managerVar.Tasks {
		list = append(list, taskItem)
	}
	managerVar.RUnlock()

	status.Tasks = list

	return status
}

func (t *Task) managerCreateErrorIsRunning() error {
	managerVar.RLock()
	defer managerVar.RUnlock()

	taskKey := t.mangerGetKeyForTask()

	taskItem, ok := managerVar.Tasks[taskKey]
	if !ok {
		return errors.New("task not found in manager")
	}

	return fmt.Errorf("task '%s' for service '%s' already started '%s' with operation '%s' - trigger '%s'", t.TaskName, t.Context.ServiceName, taskItem.StartedAt.Format(time.RFC3339), taskItem.Opeation, taskItem.TriggerType)
}

func (t *Task) mangerGetKeyForTask() string {
	return fmt.Sprintf("key___s_%s___t_%s", t.Context.ServiceName, t.Context.TaskName)
}
