package mongodb

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Restore runs mongorestore using the archive from the working directory.
func (i *MongoDB) Restore(ctx context.Context) error {
	args := i.getRestoreArgs()

	cmd := exec.Command(i.getBinMongoRestore(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
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

		return fmt.Errorf("run mongorestore: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

// RestorePipe is not supported by the MongoDB driver.
func (i *MongoDB) RestorePipe(_ context.Context, filename string) (io.WriteCloser, func() error, error) {
	return nil, func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedRestorePipe reports whether restoring from a stream is supported.
func (i *MongoDB) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}
