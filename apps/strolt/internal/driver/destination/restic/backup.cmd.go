package restic

import (
	gocontext "context"
	"fmt"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
)

// backupArgs builds the argument list of the restic backup command. The binary
// versions are passed in because collecting them runs restic itself.
func (i *Restic) backupArgs(ctx context.Context, filename string, isPipe bool, binaryVersions []interfaces.DriverBinaryVersion) []string {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "backup", "--host", "strolt_host")

	{
		for _, tag := range ctx.Tags {
			args = append(args, "--tag", tag)
		}

		for _, tag := range binaryVersions {
			args = append(args, "--tag", fmt.Sprintf("%s=%s", tag.Name, tag.Version))
		}

		args = append(args, "--tag", fmt.Sprintf("%s=%s", ldflags.GetBinaryName(), ldflags.GetVersion()))
	}

	args = append(args, "--tag", fmt.Sprintf("stroltStartedAt=%d", ctx.Operation.Time.Start.Unix()))

	args = append(args, i.getBackupFlags()...)

	if isPipe {
		args = append(args, "--stdin", "--stdin-filename="+filename)
	} else {
		args = append(args, ".")
	}

	// Appended last so a configured flag overrides the ones built above.
	args = append(args, i.config.ExtraBackupArgs...)

	return args
}

func (i *Restic) backupCmd(ctx context.Context, filename string, isPipe bool) (*exec.Cmd, error) {
	binaryVersions, err := i.BinaryVersion()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(gocontext.Background(), i.getBin(), i.backupArgs(ctx, filename, isPipe, binaryVersions)...) //nolint:gosec // restic binary path and flags come from validated config
	cmd.Dir = ctx.WorkDir

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	return cmd, nil
}
