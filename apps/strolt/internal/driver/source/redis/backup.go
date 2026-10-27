package redis

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Backup runs redis-cli --rdb and stores the RDB file in the working directory.
func (i *Redis) Backup(ctx context.Context) error {
	cmd := exec.Command(i.getBinRedisCli(), i.getBackupArgs(ArchiveFileName)...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
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

		return fmt.Errorf("run redis-cli: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return i.validateRDBArtifact(ctx.WorkDir + "/" + i.getFileName())
}

// validateRDBArtifact rejects an artifact that is not an RDB file: a server
// that answers an error string over the protocol still lets redis-cli exit 0,
// so the magic header is the real success signal. The bogus file is removed
// so no invalid artifact is ever handed to a destination.
func (i *Redis) validateRDBArtifact(filename string) error {
	file, err := os.Open(filename) //nolint:gosec // path is assembled from the driver workdir and a driver-owned constant
	if err != nil {
		return fmt.Errorf("open rdb artifact: %w", err)
	}
	defer func() { _ = file.Close() }()

	magic := make([]byte, len(rdbMagic))
	if _, err := io.ReadFull(file, magic); err != nil {
		_ = os.Remove(filename)

		return fmt.Errorf("invalid RDB file: %w", err)
	}

	if string(magic) != rdbMagic {
		_ = os.Remove(filename)

		return errors.New("invalid RDB file: bad magic")
	}

	return nil
}

// IsSupportedBackupPipe reports whether backing up as a stream is supported.
func (i *Redis) IsSupportedBackupPipe(_ context.Context) bool {
	return true
}
