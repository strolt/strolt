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

type Config struct {
	Path string `yaml:"path"`
}

type Env struct{}

type Local struct {
	logger *logger.Logger
	config Config
	env    Env
}

func New() *Local {
	return &Local{}
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
		return err
	}

	return nil
}

func validateEnv(_ Env) error {
	return nil
}

func (i *Local) Backup(_ context.Context) error {
	// if err := copy.Copy(i.config.Path, ctx.WorkDir); err != nil {
	// 	return err
	// }
	return nil
}

func (i *Local) BackupPipe(_ context.Context) (io.ReadCloser, string, func() error, error) {
	stat, err := os.Stat(i.config.Path)
	if err != nil {
		return nil, "", nil, err
	}

	if stat.IsDir() {
		return nil, "", nil, fmt.Errorf("'%s' is not file", i.config.Path)
	}

	reader, err := os.Open(i.config.Path)
	if err != nil {
		return nil, "", nil, err
	}

	// stringReader := strings.NewReader("shiny!")
	// stringReadCloser := io.NopCloser(stringReader)
	// filename := fmt.Sprintf("%d.txt", time.Now().UnixNano())

	return reader, stat.Name(), func() error { return nil }, nil
}

func (i *Local) IsSupportedBackupPipe(_ context.Context) bool {
	stat, err := os.Stat(i.config.Path)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

func (i *Local) Restore(_ context.Context) error {
	// if err := copy.Copy(ctx.WorkDir, i.config.Path); err != nil {
	// 	return err
	// }
	return nil
}

func (i *Local) RestorePipe(_ context.Context, filename string) (io.WriteCloser, func() error, error) {
	return nil, func() error { return nil }, errors.New("not support pipe")
}

func (i *Local) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}

func (i *Local) IsEmpty() (bool, error) {
	entries, err := os.ReadDir(i.config.Path)
	if err != nil {
		return false, err
	}

	isEmpty := len(entries) == 0

	return isEmpty, nil
}

func (i *Local) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	return []interfaces.DriverBinaryVersion{}, nil
}
