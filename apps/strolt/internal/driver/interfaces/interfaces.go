// Package interfaces defines the contracts implemented by source, destination, and notification drivers.
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

// Snapshot describes a single backup snapshot stored in a destination.
type Snapshot struct {
	Time    time.Time `json:"time"`
	ID      string    `json:"id"`
	ShortID string    `json:"shortId,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
	Paths   []string  `json:"paths,omitempty"`
} //	@name	Snapshot

// Stats aggregates size and count information for stored snapshots.
type Stats struct {
	TotalSize      uint64 `json:"totalSize"`
	TotalFileCount uint64 `json:"totalFileCount"`
	SnapshotsCount int    `json:"snapshotsCount"`
} //	@name	Stats

// FormattedStats extends Stats with a human-readable total size.
type FormattedStats struct {
	Stats

	TotalSizeFormatted string `json:"totalSizeFormatted"`
} //	@name	FormattedStats

// Convert returns the stats with a human-readable total size attached.
func (s *Stats) Convert() FormattedStats {
	return FormattedStats{
		Stats:              *s,
		TotalSizeFormatted: humanize.Bytes(s.TotalSize),
	}
}

// GetID returns the short snapshot ID when available, otherwise the full ID.
func (s *Snapshot) GetID() string {
	if s.ShortID != "" {
		return s.ShortID
	}

	return s.ID
}

// DriverSourceInterface is the contract implemented by backup source drivers.
type DriverSourceInterface interface { //nolint:interfacebloat // the driver contract intentionally covers the full source lifecycle
	SetLogger(l *logger.Logger)
	SetConfig(config any) error
	SetEnv(env any) error

	Backup(ctx context.Context) error
	BackupPipe(ctx context.Context) (reader io.ReadCloser, filename string, wait func() error, err error)
	IsSupportedBackupPipe(ctx context.Context) bool

	Restore(ctx context.Context) error
	RestorePipe(ctx context.Context, filename string) (writer io.WriteCloser, wait func() error, err error)
	IsSupportedRestorePipe(ctx context.Context) bool

	IsEmpty() (bool, error)
	BinaryVersion() ([]DriverBinaryVersion, error)
}

// DriverDestinationInterface is the contract implemented by backup destination drivers.
type DriverDestinationInterface interface { //nolint:interfacebloat // the driver contract intentionally covers the full destination lifecycle
	SetLogger(l *logger.Logger)
	SetConfig(config any) error
	SetEnv(env any) error
	SetTaskName(taskName string)
	SetDriverName(driverName string)

	Backup(ctx context.Context) (sctxt.BackupOutput, error)
	BackupPipe(ctx context.Context, filename string) (writer io.WriteCloser, wait func() error, err error)
	IsSupportedBackupPipe(ctx context.Context) bool

	Restore(ctx context.Context, snapshotName string) error
	RestorePipe(ctx context.Context, snapshotName string) (reader io.ReadCloser, filename string, wait func() error, err error)
	IsSupportedRestorePipe(ctx context.Context) bool

	Prune(ctx context.Context, isDryRun bool) ([]Snapshot, error)
	Stats() (Stats, error)
	Snapshots() ([]Snapshot, error)
	BinaryVersion() ([]DriverBinaryVersion, error)
	Init() error
}

// DriverBinaryVersion holds the name and version of a binary used by a driver.
type DriverBinaryVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// DriverNotificationInterface is the contract implemented by notification drivers.
type DriverNotificationInterface interface {
	SetLogger(l *logger.Logger)
	SetConfig(config any) error

	Send(ctx context.Context)
}

// DriverInitial is a placeholder driver implementation without configuration.
type DriverInitial struct{}
