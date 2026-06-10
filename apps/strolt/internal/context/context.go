// Package context provides the operation context shared between task drivers.
package context

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
	"github.com/strolt/strolt/shared/logger"
)

// Time holds the start and stop timestamps of an operation.
type Time struct {
	Start time.Time `json:"start"`
	Stop  time.Time `json:"stop"`
}

// Operation describes the timing and error state of a single operation phase.
type Operation struct {
	Time  Time   `json:"time"`
	Error string `json:"errorMessage"`
}

// DestinationOperation extends Operation with the backup output of a destination.
type DestinationOperation struct {
	Operation

	BackupOutput sctxt.BackupOutput `json:"backupOutput"`
}

// Context carries the state of a task operation across source and destinations.
type Context struct {
	Trigger        sctxt.TriggerType   `json:"trigger"`
	ServiceName    string              `json:"serviceName"`
	TaskName       string              `json:"taskName"`
	OpertationType sctxt.OperationType `json:"opertationType"`

	Event       sctxt.EventType                 `json:"event"`
	Operation   Operation                       `json:"operation"`
	Source      Operation                       `json:"source"`
	Destination map[string]DestinationOperation `json:"destination"`

	WorkDir      string `json:"workDir"`
	IsWorkDirTmp bool   `json:"isWorkDirTmp"`

	Tags []string `json:"tags"`

	SourceLocalPath string `json:"sourceLocalPath"`
}

// New creates a Context for the given task operation and prepares its work directory.
func New(trigger sctxt.TriggerType, serviceName string, taskName string, opertationType sctxt.OperationType, sourceLocalPath string) (Context, error) {
	ctx := Context{
		Trigger:        trigger,
		ServiceName:    serviceName,
		TaskName:       taskName,
		OpertationType: opertationType,
		Destination:    make(map[string]DestinationOperation),
		Event:          sctxt.EvOperationStart,

		SourceLocalPath: sourceLocalPath,
	}

	if err := ctx.setWorkDir(); err != nil {
		return Context{}, err
	}

	return ctx, nil
}

// Close removes the temporary work directory if one was created.
func (ctx *Context) Close() error {
	if ctx.IsWorkDirTmp {
		if err := dir.Remove(ctx.WorkDir); err != nil {
			return fmt.Errorf("remove work dir: %w", err)
		}
	}

	return nil
}

func (ctx *Context) setWorkDir() error {
	log := logger.New()
	log.WithFields(logger.Fields{
		"serviceName": ctx.ServiceName,
		"taskName":    ctx.TaskName,
		"operation":   ctx.OpertationType,
	})

	isNeedCreateWorkDir := ctx.SourceLocalPath == ""

	if isNeedCreateWorkDir {
		d := dir.New()
		d.SetServiceName(ctx.ServiceName)
		d.SetTaskName(ctx.TaskName)
		d.SetName("temp_work_dir")

		tempDirPath, err := d.CreateAsTmp()
		if err != nil {
			return fmt.Errorf("create temp work dir: %w", err)
		}

		ctx.WorkDir = tempDirPath
		ctx.IsWorkDirTmp = true
	} else {
		absPath, err := filepath.Abs(ctx.SourceLocalPath)
		if err != nil {
			return fmt.Errorf("resolve source local path: %w", err)
		}

		ctx.WorkDir = absPath
		ctx.IsWorkDirTmp = false
	}

	log.Debug(fmt.Sprintf("work dir path '%s'", ctx.WorkDir))

	return nil
}
