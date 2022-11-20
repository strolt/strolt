package interfaces

import (
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// type TemplateParams struct {
// 	Prefix  string
// 	Context context.Context
// }

type Snapshot struct {
	Time    time.Time `json:"time"`
	ID      string    `json:"id"`
	ShortID string    `json:"shortId,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
	Paths   []string  `json:"paths,omitempty"`
}

func (s *Snapshot) GetID() string {
	if s.ShortID != "" {
		return s.ShortID
	}

	return s.ID
}

type DriverSourceInterface interface {
	SetLogger(*logger.Logger)
	SetConfig(config interface{}) error
	SetEnv(env interface{}) error

	Backup(ctx context.Context) error
	Restore(ctx context.Context) error
	IsEmpty() (bool, error)
	BinaryVersion() ([]DriverBinaryVersion, error)
}

type DriverDestinationInterface interface {
	SetLogger(*logger.Logger)
	SetConfig(config interface{}) error
	SetEnv(env interface{}) error
	SetTaskName(taskName string)
	SetDriverName(driverName string)

	Backup(ctx context.Context) (sctxt.BackupOutput, error)
	Restore(ctx context.Context, snapshotName string) error
	Prune(ctx context.Context, isDryRun bool) ([]Snapshot, error)
	Stats() error
	Snapshots() ([]Snapshot, error)
	BinaryVersion() ([]DriverBinaryVersion, error)
	Init() error
}

type DriverBinaryVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DriverNotificationInterface interface {
	SetLogger(*logger.Logger)
	SetConfig(config interface{}) error

	Send(ctx context.Context)
}

type DriverInitial struct{}
