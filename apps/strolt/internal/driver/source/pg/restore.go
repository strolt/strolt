package pg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Restore restores the dump from the working directory using psql or pg_restore.
func (i *PgDump) Restore(ctx context.Context) error {
	filename, err := i.getFilenameFromBackup(ctx)
	if err != nil {
		return err
	}

	if i.isRestoreWithPSQL(filename) {
		return i.restoreWithPSQLCopy(ctx, filename)
	}

	return i.restoreWithPgRestoreCopy(ctx, filename)
}

func (i *PgDump) getFilenameFromBackup(ctx context.Context) (string, error) {
	files, err := os.ReadDir(ctx.WorkDir)
	if err != nil {
		return "", fmt.Errorf("read work directory: %w", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), FileNamePrefix) {
			return file.Name(), nil
		}
	}

	return "", errors.New("not found dump")
}

// RestorePipe returns a writer that streams the dump into psql or pg_restore.
func (i *PgDump) RestorePipe(ctx context.Context, filename string) (io.WriteCloser, func() error, error) {
	if i.isRestoreWithPSQL(filename) {
		return i.restoreWithPSQLPipe(ctx, filename)
	}

	return i.restoreWithPgRestorePipe(ctx, filename)
}

// IsSupportedRestorePipe reports whether restoring from a stream is supported.
func (i *PgDump) IsSupportedRestorePipe(_ context.Context) bool {
	return i.config.Format != FormatDirectory
}
