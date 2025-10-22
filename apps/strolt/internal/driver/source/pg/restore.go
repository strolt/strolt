package pg

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

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
		return "", err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), FileNamePrefix) {
			return file.Name(), nil
		}
	}

	return "", errors.New("not found dump")
}

func (i *PgDump) RestorePipe(ctx context.Context, filename string) (io.WriteCloser, func() error, error) {
	if i.isRestoreWithPSQL(filename) {
		return i.restoreWithPSQLPipe(ctx, filename)
	}

	return i.restoreWithPgRestorePipe(ctx, filename)
}

func (i *PgDump) IsSupportedRestorePipe(_ context.Context) bool {
	return i.config.Format != FormatDirectory
}
