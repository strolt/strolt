// Package restic implements a destination driver backed by the restic backup tool.
package restic

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/shared/logger"

	"gopkg.in/yaml.v3"
)

// Restic is a destination driver that stores snapshots in a restic repository.
type Restic struct {
	driverName string
	taskName   string
	logger     *logger.Logger
	config     Config
	env        Env
}

// New creates a new Restic driver instance.
func New() *Restic {
	return &Restic{}
}

// SetTaskName sets the task name for the driver.
func (i *Restic) SetTaskName(taskName string) {
	i.taskName = taskName
}

// SetDriverName sets the driver name.
func (i *Restic) SetDriverName(driverName string) {
	i.driverName = driverName
}

// SetConfig parses and validates the driver configuration.
func (i *Restic) SetConfig(config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return i.validateConfig()
}

// SetEnv parses and validates the driver environment variables.
func (i *Restic) SetEnv(env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}

	return i.validateEnv()
}

// SetLogger sets the logger for the driver.
func (i *Restic) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func startCmd(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputList := strings.Split(string(output), "\n")

		if len(outputList) > 1 {
			return nil, errors.New(outputList[0])
		}

		return nil, fmt.Errorf("run restic command: %w", err)
	}

	return output, nil
}
