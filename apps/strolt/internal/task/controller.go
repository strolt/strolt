package task

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

type ControllerOperationType string

const (
	COBackupSource      ControllerOperationType = "BACKUP_SOURCE"
	COBackupDestination ControllerOperationType = "BACKUP_DESTINATION"

	CORestoreDestination ControllerOperationType = "RESTORE_DESTINATION"
	CORestoreSource      ControllerOperationType = "RESTORE_SOURCE"

	COPruneDestination ControllerOperationType = "PRUNE_DESTINATION"

	COFetchSnapshots ControllerOperationType = "FETCH_SNAPSHOTS"

	COFetchStats ControllerOperationType = "FETCH_STATS"
)

type controller struct {
	sync.Mutex
	Operations []ControllerOperation
}

type ControllerOperation struct {
	ServiceName     string                  `json:"serviceName"`
	TaskName        string                  `json:"taskName"`
	DestinationName string                  `json:"destinationName,omitempty"`
	Operation       ControllerOperationType `json:"operation"`
	StartedAt       time.Time               `json:"startedAt"`
}

var (
	tController = controller{
		Operations: []ControllerOperation{},
	}
)

func GetOperations() []ControllerOperation {
	return tController.Operations
}

func (operation ControllerOperation) IsWorking() bool {
	for _, cOperation := range tController.Operations {
		if cOperation.ServiceName == operation.ServiceName && cOperation.TaskName == operation.TaskName {
			return true
		}
	}

	return false
}

func (operation ControllerOperation) Start() error {
	tController.Lock()

	if operation.IsWorking() {
		return errors.Wrapf(errorAlreadyWorking(operation.ServiceName, operation.TaskName), "error start")
	}

	operation.StartedAt = time.Now()
	isExistsOperation, existsOperationIndex := operation.find()

	if isExistsOperation {
		tController.Operations[existsOperationIndex] = operation
	} else {
		tController.Operations = append(tController.Operations, operation)
	}

	tController.Unlock()

	return nil
}

func (operation ControllerOperation) find() (bool, int) {
	isExistsOperation := false
	existsOperationIndex := 0

	for i, cOperation := range tController.Operations {
		if cOperation.ServiceName == operation.ServiceName && cOperation.TaskName == operation.TaskName {
			isExistsOperation = true
			existsOperationIndex = i

			break
		}
	}

	return isExistsOperation, existsOperationIndex
}

func (operation ControllerOperation) Stop() {
	tController.Lock()

	isExistsOperation, existsOperationIndex := operation.find()

	if isExistsOperation {
		tController.Operations = append(tController.Operations[:existsOperationIndex], tController.Operations[existsOperationIndex+1:]...)
	}

	tController.Unlock()
}
