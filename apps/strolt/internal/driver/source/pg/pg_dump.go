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

	cmd := exec.Command(i.getBinPgDump(), args...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	return cmd
}

// Backup runs pg_dump and stores the dump in the working directory.
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

		return fmt.Errorf("run pg_dump: %w", err)
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

// BackupPipe runs pg_dump and returns its stdout as a backup stream.
func (i *PgDump) BackupPipe(ctx context.Context) (io.ReadCloser, string, func() error, error) {
	cmd := i.backup(ctx, true)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", nil, fmt.Errorf("get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, "", nil, fmt.Errorf("start pg_dump: %w", err)
	}

	return pipe, i.getFileName(), cmd.Wait, nil
}

// IsSupportedBackupPipe reports whether backing up as a stream is supported.
func (i *PgDump) IsSupportedBackupPipe(_ context.Context) bool {
	return i.config.Format != FormatDirectory
}
