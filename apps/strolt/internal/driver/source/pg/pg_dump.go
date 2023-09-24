package pg

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *PgDump) Backup(ctx context.Context) error {
	args := i.getBackupArgs()
	args = append(args, "--file="+i.getFileName())
	cmd := exec.Command(i.getBinPgDump(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

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

func (i *PgDump) BackupPipe(_ context.Context) (io.ReadCloser, string, error) {
	return nil, "", errors.New("not support pipe")
}

func (i *PgDump) IsSupportedBackupPipe(_ context.Context) bool {
	return false
}
