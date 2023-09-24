package local

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/strolt/strolt/shared/logger"

	"github.com/google/uuid"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Path string `yaml:"path"`
}

type Env struct{}

type Local struct {
	taskName   string
	driverName string
	logger     *logger.Logger
	config     Config
	env        Env
}

func New() *Local {
	return &Local{}
}

func (i *Local) Init() error {
	return nil
}

func (i *Local) SetTaskName(taskName string) {
	i.taskName = taskName
}

func (i *Local) SetDriverName(driverName string) {
	i.driverName = driverName
}

func (i *Local) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return err
	}

	return validateConfig(i.config)
}

func (i *Local) SetEnv(env interface{}) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return err
	}

	return validateEnv(i.env)
}

func (i *Local) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func validateConfig(config Config) error {
	if config.Path == "" {
		return fmt.Errorf("not found field 'path' in config")
	}

	_, err := os.Stat(config.Path)
	if err != nil {
		if !os.IsExist(err) {
			if err := os.MkdirAll(config.Path, 0o700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func validateEnv(_ Env) error {
	return nil
}

func (i *Local) Backup(ctx context.Context) (sctxt.BackupOutput, error) {
	snapshotName := uuid.New().String()
	if err := copy.Copy(ctx.WorkDir, path.Join(i.config.Path, snapshotName)); err != nil {
		return sctxt.BackupOutput{}, err
	}

	return sctxt.BackupOutput{}, nil
}

func (i *Local) BackupPipe(ctx context.Context, filename string) (io.WriteCloser, error) {
	snapshotName := uuid.New().String()

	dirpath := path.Join(i.config.Path, snapshotName)
	if err := os.MkdirAll(dirpath, 0777); err != nil { //nolint:gomnd
		return nil, err
	}

	filepath := path.Join(dirpath, filename)

	return os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644) //nolint:gomnd
}

func (i *Local) IsSupportedBackupPipe(ctx context.Context) bool {
	return true
}

func (i *Local) Restore(ctx context.Context, snapshotName string) error {
	return copy.Copy(path.Join(i.config.Path, snapshotName), ctx.WorkDir)
}

func (i *Local) RestorePipe(ctx context.Context, snapshotName string) error {
	return errors.New("not support pipe")
}

func (i *Local) IsSupportedRestorePipe(ctx context.Context) bool {
	return false
}

func (i *Local) Prune(_ context.Context, isDryRun bool) ([]interfaces.Snapshot, error) {
	i.logger.Debug("prune")

	if isDryRun {
		return []interfaces.Snapshot{}, fmt.Errorf("dry run not supported")
	}

	return []interfaces.Snapshot{}, nil
}

func (i *Local) Stats() (interfaces.Stats, error) {
	i.logger.Debug("stats")
	return interfaces.Stats{}, nil
}

func (i *Local) Snapshots() ([]interfaces.Snapshot, error) {
	entries, err := os.ReadDir(i.config.Path)
	if err != nil {
		return nil, err
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

func (i *Local) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	return []interfaces.DriverBinaryVersion{}, nil
}
