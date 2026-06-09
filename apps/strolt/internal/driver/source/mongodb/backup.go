package mongodb

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Backup runs mongodump and stores the archive in the working directory.
func (i *MongoDB) Backup(ctx context.Context) error {
	args := i.getBackupArgs()

	cmd := exec.Command(i.getBinMongoDump(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
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

		return fmt.Errorf("run mongodump: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

// BackupPipe is not supported by the MongoDB driver.
func (i *MongoDB) BackupPipe(_ context.Context) (io.ReadCloser, string, func() error, error) {
	return nil, "", func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedBackupPipe reports whether backing up as a stream is supported.
func (i *MongoDB) IsSupportedBackupPipe(_ context.Context) bool {
	return false
}
