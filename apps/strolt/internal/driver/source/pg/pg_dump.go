package pg

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *PgDump) backup(ctx context.Context, isPipe bool) *exec.Cmd {
	args := i.getBackupArgs()

	if !isPipe {
		args = append(args, "--file="+i.getFileName())
	}

	cmd := exec.Command(i.getBinPgDump(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	return cmd
}

func (i *PgDump) Backup(ctx context.Context) error {
	cmd := i.backup(ctx, false)

	outputByte, err := cmd.CombinedOutput()
	outputString := string(outputByte)
	arr := strings.Split(outputString, "\n")

	if err != nil {
		i.logger.Error(outputString)

		if len(arr) > 0 {
			lastMessage := arr[len(arr)-1]
			if lastMessage == "" && len(arr) > 1 {
				lastMessage = arr[len(arr)-2]
			}

			return fmt.Errorf("%w (%v)", err, lastMessage)
		}

		return err
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

func (i *PgDump) BackupPipe(ctx context.Context) (io.ReadCloser, string, func() error, error) {
	cmd := i.backup(ctx, true)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, "", nil, err
	}

	return pipe, i.getFileName(), cmd.Wait, err
}

func (i *PgDump) IsSupportedBackupPipe(_ context.Context) bool {
	return i.config.Format != FormatDirectory
}
