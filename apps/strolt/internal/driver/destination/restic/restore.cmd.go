package restic

import (
	gocontext "context"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// getRestoreTarget returns the directory restic restores into.
//
// By default that is the task work directory, which for a local source is the
// source path itself, so the snapshot lands where the task expects it. A
// configured restore-target redirects restic to a plain directory instead —
// useful when the source is a mountpoint that cannot serve the concurrent
// random writes of the restorer (FUSE). strolt does not move that directory
// into the source afterwards, hence the warning.
func (i *Restic) getRestoreTarget(ctx context.Context) string {
	if i.config.RestoreTarget == "" {
		return ctx.WorkDir
	}

	i.logger.Warnf(
		"restoring into the configured restore-target '%s' instead of '%s'; strolt does not move the restored data into the source",
		i.config.RestoreTarget, ctx.WorkDir,
	)

	return i.config.RestoreTarget
}

// restoreArgs builds the argument list of the restic restore or dump command.
func (i *Restic) restoreArgs(ctx context.Context, snapshotID string, path string, isPipe bool) []string {
	var args []string
	args = append(args, i.getGlobalFlags()...)

	if isPipe {
		return append(args, "dump", snapshotID, path)
	}

	args = append(args, "restore", snapshotID, "--target", i.getRestoreTarget(ctx))

	// Appended last so a configured flag overrides the ones built above.
	args = append(args, i.config.ExtraRestoreArgs...)

	return args
}

func (i *Restic) restoreCmd(ctx context.Context, snapshotID string, path string, isPipe bool) (*exec.Cmd, error) {
	cmd := exec.CommandContext(gocontext.Background(), i.getBin(), i.restoreArgs(ctx, snapshotID, path, isPipe)...) //nolint:gosec // restic binary path and flags come from validated config

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	return cmd, nil
}
