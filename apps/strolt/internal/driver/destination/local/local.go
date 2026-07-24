// Package local implements a destination driver that stores backups in a local directory.
package local

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/strolt/strolt/shared/logger"

	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v3"
)

// Config describes the local destination driver configuration.
type Config struct {
	Path string `yaml:"path"`
}

// Env describes the local destination driver environment variables.
type Env struct{}

// Local is a destination driver that stores snapshots in a local directory.
type Local struct {
	taskName   string
	driverName string
	logger     *logger.Logger
	config     Config
	env        Env
}

// New creates a new Local driver instance.
func New() *Local {
	return &Local{}
}

// Init initializes the driver.
func (i *Local) Init() error {
	return nil
}

// SetTaskName sets the task name for the driver.
func (i *Local) SetTaskName(taskName string) {
	i.taskName = taskName
}

// SetDriverName sets the driver name.
func (i *Local) SetDriverName(driverName string) {
	i.driverName = driverName
}

// SetConfig parses and validates the driver configuration.
func (i *Local) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return validateConfig(i.config)
}

// SetEnv parses and validates the driver environment variables.
func (i *Local) SetEnv(env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}

	return validateEnv(i.env)
}

// SetLogger sets the logger for the driver.
func (i *Local) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func validateConfig(config Config) error {
	if config.Path == "" {
		return errors.New("not found field 'path' in config")
	}

	_, err := os.Stat(config.Path)
	if err != nil {
		if !os.IsExist(err) {
			if err := os.MkdirAll(config.Path, 0o700); err != nil {
				return fmt.Errorf("create destination directory: %w", err)
			}
		} else {
			return fmt.Errorf("stat destination directory: %w", err)
		}
	}

	return nil
}

func validateEnv(_ Env) error {
	return nil
}

// Backup copies the working directory into a new snapshot directory.
func (i *Local) Backup(ctx context.Context) (sctxt.BackupOutput, error) {
	snapshotName := uuid.New().String()
	if err := copy.Copy(ctx.WorkDir, path.Join(i.config.Path, snapshotName)); err != nil {
		return sctxt.BackupOutput{}, fmt.Errorf("copy workdir to snapshot: %w", err)
	}

	return sctxt.BackupOutput{}, nil
}

// BackupPipe returns a writer that stores the piped backup in a new snapshot directory.
// The local driver produces no restic-style summary, so its wait func returns an
// empty BackupOutput; the streamed byte count is filled in by the caller.
func (i *Local) BackupPipe(ctx context.Context, filename string) (io.WriteCloser, func() (sctxt.BackupOutput, error), error) {
	snapshotName := uuid.New().String()

	dirpath := path.Join(i.config.Path, snapshotName)
	if err := os.MkdirAll(dirpath, 0o750); err != nil { //nolint:mnd
		return nil, nil, fmt.Errorf("create snapshot directory: %w", err)
	}

	filepath := path.Join(dirpath, filename)

	writer, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:mnd,gosec // path is built from the configured destination directory
	if err != nil {
		return nil, nil, fmt.Errorf("open snapshot file: %w", err)
	}

	return writer, func() (sctxt.BackupOutput, error) { return sctxt.BackupOutput{}, nil }, nil
}

// IsSupportedBackupPipe reports whether piped backups are supported.
func (i *Local) IsSupportedBackupPipe(ctx context.Context) bool {
	return true
}

// Restore copies a snapshot directory back into the working directory.
func (i *Local) Restore(ctx context.Context, snapshotName string) error {
	if err := copy.Copy(path.Join(i.config.Path, snapshotName), ctx.WorkDir); err != nil {
		return fmt.Errorf("copy snapshot to workdir: %w", err)
	}

	return nil
}

// RestorePipe returns a reader for the single file stored in a snapshot.
func (i *Local) RestorePipe(ctx context.Context, snapshotName string) (io.ReadCloser, string, func() error, error) {
	list, err := i.ls(snapshotName)
	if err != nil {
		return nil, "", nil, err
	}

	if len(list) != 1 || list[0].IsDir() {
		return nil, "", nil, errors.New("not supported")
	}

	filename := list[0].Name()

	reader, err := os.Open(path.Join(i.config.Path, snapshotName, filename)) //nolint:gosec // path is built from the configured destination directory
	if err != nil {
		return nil, "", nil, fmt.Errorf("open snapshot file: %w", err)
	}

	return reader, filename, func() error { return nil }, nil
}

// IsSupportedRestorePipe reports whether piped restores are supported.
func (i *Local) IsSupportedRestorePipe(ctx context.Context) bool {
	return true
}

// Prune removes old snapshots; dry run is not supported.
func (i *Local) Prune(_ context.Context, isDryRun bool) ([]interfaces.Snapshot, error) {
	i.logger.Debug("prune")

	if isDryRun {
		return []interfaces.Snapshot{}, errors.New("dry run not supported")
	}

	return []interfaces.Snapshot{}, nil
}

// Forget removes a single snapshot directory from the destination. The
// snapshot ID is matched against the stored snapshots first: it comes from user
// input and is joined into a filesystem path, so an unchecked value could point
// outside the destination directory.
func (i *Local) Forget(snapshotID string) error {
	snapshots, err := i.Snapshots()
	if err != nil {
		return err
	}

	isKnown := slices.ContainsFunc(snapshots, func(snapshot interfaces.Snapshot) bool {
		return snapshot.ID == snapshotID
	})

	if !isKnown {
		return fmt.Errorf("snapshot '%s' not found in destination", snapshotID)
	}

	if err := os.RemoveAll(path.Join(i.config.Path, snapshotID)); err != nil {
		return fmt.Errorf("remove snapshot directory: %w", err)
	}

	return nil
}

// Unlock does nothing: the local destination stores snapshots as plain
// directories and never takes repository locks.
func (i *Local) Unlock(_ bool) error {
	i.logger.Debug("unlock: nothing to do for the local destination")

	return nil
}

// Stats returns statistics for the destination.
func (i *Local) Stats() (interfaces.Stats, error) {
	i.logger.Debug("stats")
	return interfaces.Stats{}, nil
}

// Snapshots lists the snapshots stored in the destination directory.
func (i *Local) Snapshots() ([]interfaces.Snapshot, error) {
	entries, err := os.ReadDir(i.config.Path)
	if err != nil {
		return nil, fmt.Errorf("read destination directory: %w", err)
	}

	var snapshots []interfaces.Snapshot

	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, interfaces.Snapshot{
				ID:   entry.Name(),
				Time: time.Now(),
			})
		}
	}

	return snapshots, nil
}

// BinaryVersion returns the versions of external binaries used by the driver.
func (i *Local) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	return []interfaces.DriverBinaryVersion{}, nil
}

func (i *Local) ls(snapshotName string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(path.Join(i.config.Path, snapshotName))
	if err != nil {
		return nil, fmt.Errorf("read snapshot directory: %w", err)
	}

	return entries, nil
}
