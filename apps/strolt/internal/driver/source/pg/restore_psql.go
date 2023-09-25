package pg

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *PgDump) isRestoreWithPSQL(filename string) bool {
	return strings.HasSuffix(filename, ".sql")
}

func (i *PgDump) restoreWithPSQLCmd(ctx context.Context, filename string, isPipe bool) *exec.Cmd {
	i.logger.Info("restore with psql")

	args := i.getCommonArgs()

	if !isPipe {
		args = append(args, "--file="+filename)
	}

	cmd := exec.Command(i.getBinPSQL(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	return cmd
}

func (i *PgDump) restoreWithPSQLCopy(ctx context.Context, filename string) error {
	cmd := i.restoreWithPSQLCmd(ctx, filename, false)

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

func (i *PgDump) restoreWithPSQLPipe(ctx context.Context, filename string) (io.WriteCloser, func() error, error) {
	cmd := i.restoreWithPSQLCmd(ctx, filename, true)

	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return writer, cmd.Wait, err
}
