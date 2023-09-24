package interfaces

import (
	"io"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
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
} // @name Snapshot

type Stats struct {
	TotalSize      uint64 `json:"totalSize"`
	TotalFileCount uint64 `json:"totalFileCount"`
	SnapshotsCount int    `json:"snapshotsCount"`
} // @name Stats

type FormattedStats struct {
	Stats
	TotalSizeFormatted string `json:"totalSizeFormatted"`
} // @name FormattedStats

func (s *Stats) Convert() FormattedStats {
	return FormattedStats{
		Stats:              *s,
		TotalSizeFormatted: humanize.Bytes(s.TotalSize),
	}
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
	BackupPipe(ctx context.Context) (io.ReadCloser, string, error)
	IsSupportedBackupPipe(ctx context.Context) bool

	Restore(ctx context.Context) error
	RestorePipe(ctx context.Context) error
	IsSupportedRestorePipe(ctx context.Context) bool

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
	BackupPipe(ctx context.Context, filename string) (io.WriteCloser, error)
	IsSupportedBackupPipe(ctx context.Context) bool

	Restore(ctx context.Context, snapshotName string) error
	RestorePipe(ctx context.Context, snapshotName string) error
	IsSupportedRestorePipe(ctx context.Context) bool

	Prune(ctx context.Context, isDryRun bool) ([]Snapshot, error)
	Stats() (Stats, error)
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
