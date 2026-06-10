package mysql

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Backup runs mysqldump and stores the dump in the working directory.
func (i *MySQL) Backup(ctx context.Context) error {
	args := i.getBackupArgs()
	cmd := exec.Command(i.getBinMySQLDump(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
	cmd.Dir = ctx.WorkDir

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

		return fmt.Errorf("run mysqldump: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

// BackupPipe is not supported by the MySQL driver.
func (i *MySQL) BackupPipe(_ context.Context) (io.ReadCloser, string, func() error, error) {
	return nil, "", func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedBackupPipe reports whether backing up as a stream is supported.
func (i *MySQL) IsSupportedBackupPipe(_ context.Context) bool {
	return false
}
