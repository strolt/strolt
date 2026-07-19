package mysql

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Restore runs the mysql client to load the dump from the working directory.
func (i *MySQL) Restore(ctx context.Context) error {
	dumpPath := filepath.Join(ctx.WorkDir, i.getFileName())

	dump, err := os.Open(dumpPath) //nolint:gosec // path is built from the driver's own dump file name inside the work dir
	if err != nil {
		return fmt.Errorf("open dump file: %w", err)
	}

	defer func() {
		_ = dump.Close()
	}()

	args := i.getRestoreArgs()
	cmd := exec.Command(i.getBinMySQL(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
	cmd.Dir = ctx.WorkDir
	// Feed the dump on stdin so the client runs in batch mode and aborts on the
	// first failing statement with a non-zero exit code.
	cmd.Stdin = dump

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

		return fmt.Errorf("run mysql: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

// RestorePipe is not supported by the MySQL driver.
func (i *MySQL) RestorePipe(_ context.Context, filename string) (io.WriteCloser, func() error, error) {
	return nil, func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedRestorePipe reports whether restoring from a stream is supported.
func (i *MySQL) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}
