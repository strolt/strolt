package restic

import (
	gocontext "context"
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

	cmd := exec.CommandContext(gocontext.Background(), i.getBin(), args...) //nolint:gosec // restic binary path and flags come from validated config

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	return cmd, nil
}
