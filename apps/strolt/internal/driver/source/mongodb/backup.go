package mongodb

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *MongoDB) Backup(ctx context.Context) error {
	args := i.getBackupArgs()

	cmd := exec.Command(i.getBinMongoDump(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	i.logger.Debug(cmd.Path)

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

func (i *MongoDB) BackupPipe(_ context.Context) (io.ReadCloser, string, func() error, error) {
	return nil, "", func() error { return nil }, errors.New("not support pipe")
}

func (i *MongoDB) IsSupportedBackupPipe(_ context.Context) bool {
	return false
}
