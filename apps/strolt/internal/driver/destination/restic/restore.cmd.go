package restic

import (
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *Restic) restoreCmd(ctx context.Context, snapshotID string, path string, isPipe bool) (*exec.Cmd, error) {
	var args []string
	args = append(args, i.getGlobalFlags()...)

	if isPipe {
		args = append(args, "dump", snapshotID, path)
	} else {
		args = append(args, "restore", snapshotID, "--target", ctx.WorkDir)
	}

	cmd := exec.Command(i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	return cmd, nil
}
