package restic

import (
	"context"
	"fmt"
	"os/exec"
)

// Init initializes the restic repository.
func (i *Restic) Init() error {
	globalFlags := i.getGlobalFlags()

	args := make([]string, 0, len(globalFlags)+1)
	args = append(args, globalFlags...)
	args = append(args, "init")

	cmd := exec.CommandContext(context.Background(), i.getBin(), args...) //nolint:gosec // restic binary path and flags come from validated config

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
		return fmt.Errorf("restic init: %w", err)
	}

	return nil
}
