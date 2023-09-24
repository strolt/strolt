package restic

import (
	"errors"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *Restic) Restore(ctx context.Context, snapshotID string) error {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "restore", snapshotID, "--target", ctx.WorkDir)

	cmd := exec.Command(i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)
	if err != nil {
		i.logger.Error(err)
		return err
	}

	i.logger.Info(string(output))

	return nil
}

func (i *Restic) RestorePipe(ctx context.Context, snapshotName string) error {
	return errors.New("not support pipe")
}

func (i *Restic) IsSupportedRestorePipe(ctx context.Context) bool {
	return false
}
