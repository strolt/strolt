package restic

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/logger"

	"gopkg.in/yaml.v3"
)

type Restic struct {
	driverName string
	taskName   string
	logger     *logger.Logger
	config     Config
	env        Env
}

func New() *Restic {
	return &Restic{}
}

func (i *Restic) SetTaskName(taskName string) {
	i.taskName = taskName
}

func (i *Restic) SetDriverName(driverName string) {
	i.driverName = driverName
}

func (i *Restic) SetConfig(config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return err
	}

	if err := i.validateConfig(); err != nil {
		return err
	}

	return nil
}

func (i *Restic) SetEnv(env interface{}) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return err
	}

	if err := i.validateEnv(); err != nil {
		return err
	}

	return nil
}

func (i *Restic) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func (i *Restic) Stats() error {
	i.logger.Debug("stats")
	return nil
}

func startCmd(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()

	if err != nil {
		outputList := strings.Split(string(output), "\n")

		if len(outputList) > 1 {
			return nil, fmt.Errorf(outputList[0])
		}

		return nil, err
	}

	return output, nil
}
