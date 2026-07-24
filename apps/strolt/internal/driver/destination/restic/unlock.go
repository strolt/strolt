package restic

import (
	"context"
	"os/exec"
)

// Unlock removes locks from the repository. By default restic only removes
// stale locks, i.e. locks whose owning process is gone; isRemoveAll also
// removes the locks of operations that are still running.
func (i *Restic) Unlock(isRemoveAll bool) error {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "unlock")

	if isRemoveAll {
		args = append(args, "--remove-all")
	}

	cmd := exec.CommandContext(context.Background(), i.getBin(), args...) //nolint:gosec // restic binary path and flags come from validated config

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)

	i.logger.Debug(string(output))

	if err != nil {
		i.logger.Error(err)
		return err
	}

	return nil
}
