// Package local provides a source driver for local filesystem paths.
package local

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/logger"

	"gopkg.in/yaml.v3"
)

// Config describes the local source driver configuration.
type Config struct {
	Path string `yaml:"path"`
}

// Env describes the local source driver environment variables.
type Env struct{}

// Local is a source driver that works with local filesystem paths.
type Local struct {
	logger *logger.Logger
	config Config
	env    Env
}

// New creates a new Local driver instance.
func New() *Local {
	return &Local{}
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

// SetLogger sets the logger used by the driver.
func (i *Local) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func validateConfig(config Config) error {
	if config.Path == "" {
		return errors.New("not found field 'path' in config")
	}

	_, err := os.Stat(config.Path)
	if err != nil {
		return fmt.Errorf("stat path: %w", err)
	}

	return nil
}

func validateEnv(_ Env) error {
	return nil
}

// Backup performs a backup of the configured local path.
//
// It is intentionally a no-op: for a local source the task work directory is the
// source path itself (context.setWorkDir sets WorkDir = abs(SourceLocalPath)),
// so the destination driver reads directly from the source path and there is
// nothing to copy into a separate work directory. Copying here (WorkDir ==
// config.Path) would copy the directory onto itself.
func (i *Local) Backup(_ context.Context) error {
	return nil
}

// BackupPipe opens the configured file and returns it as a backup stream.
func (i *Local) BackupPipe(_ context.Context) (io.ReadCloser, string, func() error, error) {
	stat, err := os.Stat(i.config.Path)
	if err != nil {
		return nil, "", nil, fmt.Errorf("stat path: %w", err)
	}

	if stat.IsDir() {
		return nil, "", nil, fmt.Errorf("'%s' is not file", i.config.Path)
	}

	reader, err := os.Open(i.config.Path)
	if err != nil {
		return nil, "", nil, fmt.Errorf("open file: %w", err)
	}

	// stringReader := strings.NewReader("shiny!")
	// stringReadCloser := io.NopCloser(stringReader)
	// filename := fmt.Sprintf("%d.txt", time.Now().UnixNano())

	return reader, stat.Name(), func() error { return nil }, nil
}

// IsSupportedBackupPipe reports whether the configured path can be backed up as a stream.
func (i *Local) IsSupportedBackupPipe(_ context.Context) bool {
	stat, err := os.Stat(i.config.Path)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

// Restore restores a backup to the configured local path.
//
// It is intentionally a no-op: for a local source the task work directory is the
// source path itself (context.setWorkDir sets WorkDir = abs(SourceLocalPath)),
// so the destination driver restores the snapshot directly into the source path
// (restic runs `restore --target WorkDir`). Copying here (WorkDir ==
// config.Path) would copy the directory onto itself.
func (i *Local) Restore(_ context.Context) error {
	return nil
}

// RestorePipe is not supported by the local driver.
func (i *Local) RestorePipe(_ context.Context, filename string) (io.WriteCloser, func() error, error) {
	return nil, func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedRestorePipe reports whether restoring from a stream is supported.
func (i *Local) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}

// IsEmpty reports whether the configured directory contains no entries.
func (i *Local) IsEmpty() (bool, error) {
	entries, err := os.ReadDir(i.config.Path)
	if err != nil {
		return false, fmt.Errorf("read directory: %w", err)
	}

	isEmpty := len(entries) == 0

	return isEmpty, nil
}

// BinaryVersion returns versions of external binaries used by the driver.
func (i *Local) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	return []interfaces.DriverBinaryVersion{}, nil
}
