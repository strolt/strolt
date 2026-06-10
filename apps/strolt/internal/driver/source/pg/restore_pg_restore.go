package pg

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *PgDump) restoreWithPgRestoreCmd(ctx context.Context, filename string, isPipe bool) *exec.Cmd {
	i.logger.Info("restore with pg_restore")

	args := i.getRestoreArgs()
	if !isPipe {
		args = append(args, filename)
	}

	cmd := exec.Command(i.getBinPgRestore(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	return cmd
}

func (i *PgDump) restoreWithPgRestoreCopy(ctx context.Context, filename string) error {
	cmd := i.restoreWithPgRestoreCmd(ctx, filename, false)

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

		return fmt.Errorf("run pg_restore: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

func (i *PgDump) restoreWithPgRestorePipe(ctx context.Context, filename string) (io.WriteCloser, func() error, error) {
	cmd := i.restoreWithPgRestoreCmd(ctx, filename, true)

	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("get stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start pg_restore: %w", err)
	}

	return writer, cmd.Wait, nil
}
