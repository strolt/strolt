package restic

import (
	"fmt"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
)

func (i *Restic) backupCmd(ctx context.Context, filename string, isPipe bool) (*exec.Cmd, error) {
	tags, err := i.BinaryVersion()
	if err != nil {
		return nil, err
	}

	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "backup", "--host", "strolt_host")

	{
		for _, tag := range ctx.Tags {
			args = append(args, "--tag", tag)
		}

		for _, tag := range tags {
			args = append(args, "--tag", fmt.Sprintf("%s=%s", tag.Name, tag.Version))
		}

		args = append(args, "--tag", fmt.Sprintf("%s=%s", ldflags.GetBinaryName(), ldflags.GetVersion()))
	}

	args = append(args, "--tag", fmt.Sprintf("stroltStartedAt=%d", ctx.Operation.Time.Start.Unix()))

	args = append(args, i.getBackupFlags()...)

	if isPipe {
		args = append(args, "--stdin", fmt.Sprintf("--stdin-filename=%s", filename))
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command(i.getBin(), args...)
	cmd.Dir = ctx.WorkDir

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	return cmd, nil
}
