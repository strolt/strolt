package context

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
)

type Time struct {
	Start time.Time `json:"start"`
	Stop  time.Time `json:"stop"`
}

type Operation struct {
	BackupOutput sctxt.BackupOutput `json:"backupOutput"`

	Time  Time   `json:"time"`
	Error string `json:"errorMessage"`
}

type Context struct {
	Trigger        sctxt.TriggerType   `json:"trigger"`
	ServiceName    string              `json:"serviceName"`
	TaskName       string              `json:"taskName"`
	OpertationType sctxt.OperationType `json:"opertationType"`

	Event       sctxt.EventType      `json:"event"`
	Operation   Operation            `json:"operation"`
	Source      Operation            `json:"source"`
	Destination map[string]Operation `json:"destination"`

	WorkDir      string `json:"workDir"`
	IsWorkDirTmp bool   `json:"isWorkDirTmp"`

	Tags []string `json:"tags"`

	SourceLocalPath string `json:"sourceLocalPath"`
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
			return err
		}

		ctx.WorkDir = tempDirPath
		ctx.IsWorkDirTmp = true
	} else {
		absPath, err := filepath.Abs(ctx.SourceLocalPath)
		if err != nil {
			return err
		}

		ctx.WorkDir = absPath
		ctx.IsWorkDirTmp = false
	}

	log.Debug(fmt.Sprintf("work dir path '%s'", ctx.WorkDir))

	return nil
}

func New(trigger sctxt.TriggerType, serviceName string, taskName string, opertationType sctxt.OperationType, sourceLocalPath string) (Context, error) {
	ctx := Context{
		Trigger:        trigger,
		ServiceName:    serviceName,
		TaskName:       taskName,
		OpertationType: opertationType,
		Destination:    make(map[string]Operation),

		SourceLocalPath: sourceLocalPath,
	}

	if err := ctx.setWorkDir(); err != nil {
		return Context{}, err
	}

	return ctx, nil
}

func (ctx *Context) Close() error {
	if ctx.IsWorkDirTmp {
		return dir.Remove(ctx.WorkDir)
	}

	return nil
}
