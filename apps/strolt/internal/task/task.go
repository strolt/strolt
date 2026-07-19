package task

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

// Operation holds scheduling and last-run information for a task operation.
type Operation struct {
	Schedule             string `json:"schedule"`
	LatestRun            int64  `json:"latestRun"`
	LatestDuration       int64  `json:"latestDuration"`
	IsPreviouslyLaunched bool   `json:"isPreviouslyLaunched"`
}

// Task represents a configured backup task bound to a service.
type Task struct {
	Context     context.Context   `json:"context"`
	Trigger     sctxt.TriggerType `json:"trigger"`
	ServiceName string            `json:"serviceName"`
	TaskName    string            `json:"taskName"`

	OperationBackup Operation `json:"operationBackup"`
	OperationPrune  Operation `json:"operationPrune"`

	TaskConfig              config.TaskConfig `json:"taskConfig"`
	log                     *logger.Logger
	isNotificationsDisabled bool

	// notificationWaitGroup tracks the in-flight notification goroutines of this
	// task only. It must be per-task: a package-global would let one task's Add
	// race another task's Wait, corrupting the barrier and risking a
	// "WaitGroup misuse" panic when tasks run concurrently.
	notificationWaitGroup *sync.WaitGroup
}

// TaskOperation identifies the kind of operation a task performs.
type TaskOperation string //nolint:revive // renaming the exported type would break the public API

// Supported task operations.
const (
	Backup TaskOperation = "BACKUP"
	Forget TaskOperation = "FORGET"
	Prune  TaskOperation = "PRUNE"
)

// New creates a Task for the given service and task name with a prepared context.
func New(serviceName string, taskName string, trigger sctxt.TriggerType, operationType sctxt.OperationType) (*Task, error) {
	c, err := config.GetConfigForTask(serviceName, taskName)
	if err != nil {
		return &Task{}, fmt.Errorf("get config for task: %w", err)
	}

	sourceLocalPath := ""

	if c.Source.Driver == dmanager.DriverSourceLocal {
		sourceConfig, ok := c.Source.Config.(map[string]any)
		if !ok {
			return &Task{}, fmt.Errorf("want type map[string]interface{};  got %T", c.Source.Config)
		}

		path, ok := sourceConfig["path"].(string)
		if !ok {
			return &Task{}, fmt.Errorf("want type map[string]interface{};  got %T", sourceConfig["path"])
		}

		sourceLocalPath = path
	}

	ctx, err := context.New(trigger, serviceName, taskName, operationType, sourceLocalPath)
	if err != nil {
		return &Task{}, fmt.Errorf("create task context: %w", err)
	}

	if operationType == sctxt.OpTypeBackup {
		ctx.Tags = append(ctx.Tags, c.Tags...)
		ctx.Tags = append(ctx.Tags, fmt.Sprintf("trigger=%s", ctx.Trigger))

		{
			sourceDriver, err := dmanager.GetSourceDriver(c.Source.Driver, serviceName, taskName, c.Source.Config, c.Source.Env)
			if err != nil {
				return &Task{}, fmt.Errorf("get source driver: %w", err)
			}

			sourceBinVersions, err := sourceDriver.BinaryVersion()
			if err != nil {
				return &Task{}, fmt.Errorf("get source binary version: %w", err)
			}

			for _, bin := range sourceBinVersions {
				ctx.Tags = append(ctx.Tags, fmt.Sprintf("%s=%s", bin.Name, bin.Version))
			}
		}
	}

	return &Task{
		Context:     ctx,
		Trigger:     trigger,
		ServiceName: serviceName,
		TaskName:    taskName,
		TaskConfig:  c,

		OperationBackup: Operation{
			Schedule: c.Schedule.Backup,
		},
		log: logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule),

		OperationPrune: Operation{
			Schedule: c.Schedule.Prune,
		},
	}, nil
}

// Close releases the resources held by the task context.
func (t *Task) Close() error {
	if err := t.Context.Close(); err != nil {
		return fmt.Errorf("close task context: %w", err)
	}

	return nil
}

// Clone returns a deep copy of the task created via JSON round-trip.
func (t *Task) Clone() (Task, error) {
	data, err := json.Marshal(t) //nolint:musttag // nested config.TaskConfig is yaml-tagged; the JSON round-trip only needs symmetry
	if err != nil {
		return Task{}, fmt.Errorf("marshal task: %w", err)
	}

	var tClone Task
	if err := json.Unmarshal(data, &tClone); err != nil { //nolint:musttag // nested config.TaskConfig is yaml-tagged; the JSON round-trip only needs symmetry
		return Task{}, fmt.Errorf("unmarshal task: %w", err)
	}

	return tClone, nil
}

// UpdateMetricsAfterTaskFinish updates operation metrics based on the final context event.
func (t *Task) UpdateMetricsAfterTaskFinish() {
	log := logger.New()

	if t.Context.OpertationType == sctxt.OpTypeBackup {
		if t.Context.Event == sctxt.EvOperationError {
			metrics.Operations().BackupError()
			log.Warn("updateMetrics: backup error")
		}

		if t.Context.Event == sctxt.EvOperationStop {
			log.Warn("updateMetrics: backup success")
			metrics.Operations().BackupSuccess()
		}
	}

	if t.Context.OpertationType == sctxt.OpTypePrune {
		if t.Context.Event == sctxt.EvOperationError {
			metrics.Operations().PruneError()
			log.Warn("updateMetrics: prune error")
		}

		if t.Context.Event == sctxt.EvOperationStop {
			log.Warn("updateMetrics: prune success")
			metrics.Operations().PruneSuccess()
		}
	}
}
