package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

type manager struct {
	*sync.RWMutex

	Tasks         map[string]ManagerTaskItem
	LastChangedAt time.Time
}

// ManagerTaskItem describes the manager state of a single task.
type ManagerTaskItem struct {
	ServiceName string              `json:"serviceName"`
	TaskName    string              `json:"taskName"`
	Opeation    sctxt.OperationType `json:"operation"`
	StartedAt   time.Time           `json:"startedAt"`
	LastEndedAt time.Time           `json:"lastEndedAt"`
	TriggerType sctxt.TriggerType   `json:"trigger"`
	IsRunning   bool                `json:"isRunning"`
} //	@name	ManagerTaskItem

// ManagerStatus is a snapshot of all tasks tracked by the manager.
type ManagerStatus struct {
	Tasks         []ManagerTaskItem `json:"tasks"`
	LastChangedAt string            `json:"lastChangedAt"`
} //	@name	ManagerStatus

var managerVar = manager{
	Tasks:         map[string]ManagerTaskItem{},
	LastChangedAt: time.Now(),
	RWMutex:       &sync.RWMutex{},
}

func (t *Task) managerStart(operation sctxt.OperationType) error {
	taskKey := t.mangerGetKeyForTask()

	// The running check and the running-flag update must happen inside a single
	// critical section. Splitting them (check under RLock, set under Lock) leaves
	// a TOCTOU window in which two goroutines starting the same task both observe
	// "not running" and both proceed — defeating the whole purpose of the manager.
	managerVar.Lock()
	defer managerVar.Unlock()

	if taskItem, ok := managerVar.Tasks[taskKey]; ok && taskItem.IsRunning {
		return fmt.Errorf(
			"task '%s' for service '%s' already started '%s' with operation '%s' - trigger '%s'",
			t.TaskName, t.Context.ServiceName, taskItem.StartedAt.Format(time.RFC3339), taskItem.Opeation, taskItem.TriggerType,
		)
	}

	item := ManagerTaskItem{
		ServiceName: t.Context.ServiceName,
		TaskName:    t.Context.TaskName,
		IsRunning:   true,
		StartedAt:   time.Now(),
		LastEndedAt: time.Now(),
		Opeation:    operation,
		TriggerType: t.Trigger,
	}

	if taskItem, ok := managerVar.Tasks[taskKey]; ok {
		item.LastEndedAt = taskItem.LastEndedAt
	}

	managerVar.Tasks[taskKey] = item

	managerVar.LastChangedAt = time.Now()

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

// IsRunning reports whether the task is currently running.
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

// GetLastChangedManager returns the time of the last manager state change.
func GetLastChangedManager() time.Time {
	managerVar.RLock()
	defer managerVar.RUnlock()

	return managerVar.LastChangedAt
}

// GetManagerStatus returns a snapshot of all tasks tracked by the manager.
func GetManagerStatus() ManagerStatus {
	status := ManagerStatus{}

	managerVar.RLock()
	status.LastChangedAt = managerVar.LastChangedAt.Format(time.RFC3339)

	list := make([]ManagerTaskItem, 0, len(managerVar.Tasks))
	for _, taskItem := range managerVar.Tasks {
		list = append(list, taskItem)
	}
	managerVar.RUnlock()

	status.Tasks = list

	return status
}

func (t *Task) mangerGetKeyForTask() string {
	return fmt.Sprintf("key___s_%s___t_%s", t.Context.ServiceName, t.Context.TaskName)
}
