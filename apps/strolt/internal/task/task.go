package task

import (
	"encoding/json"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

type Operation struct {
	Schedule             string
	LatestRun            int64
	LatestDuration       int64
	IsPreviouslyLaunched bool
}

//nolint:containedctx,musttag
type Task struct {
	Context     context.Context
	Trigger     sctxt.TriggerType
	ServiceName string
	TaskName    string

	OperationBackup Operation
	OperationPrune  Operation

	TaskConfig              config.TaskConfig
	log                     *logger.Logger
	isNotificationsDisabled bool
}

type TaskOperation string

const (
	Backup TaskOperation = "BACKUP"
	Forget TaskOperation = "FORGET"
	Prune  TaskOperation = "PRUNE"
)

func (t *Task) Close() error {
	return t.Context.Close()
}

func (t *Task) Clone() (Task, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return Task{}, err
	}

	var tClone Task
	if err := json.Unmarshal(data, &tClone); err != nil {
		return Task{}, err
	}

	return tClone, nil
}

func New(serviceName string, taskName string, trigger sctxt.TriggerType, operationType sctxt.OperationType) (*Task, error) {
	c, err := config.GetConfigForTask(serviceName, taskName)
	if err != nil {
		return &Task{}, err
	}

	sourceLocalPath := ""

	if c.Source.Driver == dmanager.DriverSourceLocal {
		sourceConfig, ok := c.Source.Config.(map[string]interface{})
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
		return &Task{}, err
	}

	if operationType == sctxt.OpTypeBackup {
		ctx.Tags = append(ctx.Tags, c.Tags...)
		ctx.Tags = append(ctx.Tags, fmt.Sprintf("trigger=%s", ctx.Trigger))

		{
			sourceDriver, err := dmanager.GetSourceDriver(c.Source.Driver, serviceName, taskName, c.Source.Config, c.Source.Env)
			if err != nil {
				return &Task{}, err
			}

			sourceBinVersions, err := sourceDriver.BinaryVersion()
			if err != nil {
				return &Task{}, err
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
