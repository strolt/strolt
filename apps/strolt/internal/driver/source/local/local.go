package local

import (
	"fmt"
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

func (i *Local) Restore(_ context.Context) error {
	// if err := copy.Copy(ctx.WorkDir, i.config.Path); err != nil {
	// 	return err
	// }
	return nil
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
