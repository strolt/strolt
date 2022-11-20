package restic

import (
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *Restic) Restore(ctx context.Context, snapshotID string) error {
	cmd := exec.Command(i.getBin(), "restore", snapshotID, "--target", ctx.WorkDir)

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := cmd.Output()
	if err != nil {
		i.logger.Error(err)
		return err
	}

	i.logger.Info(string(output))

	return nil
}
