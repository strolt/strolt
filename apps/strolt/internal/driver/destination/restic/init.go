package restic

import (
	"os/exec"
)

func (i *Restic) Init() error {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "init")

	cmd := exec.Command(i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	i.logger.Debug(cmd.Env)

	output, err := cmd.CombinedOutput()
	i.logger.Debug(string(output))

	if err != nil {
		return err
	}

	return nil
}
