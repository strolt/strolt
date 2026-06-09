package restic

import (
	"fmt"
	"io"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Restore restores a restic snapshot into the working directory.
func (i *Restic) Restore(ctx context.Context, snapshotID string) error {
	cmd, err := i.restoreCmd(ctx, snapshotID, "", false)
	if err != nil {
		i.logger.Error(err)
		return err
	}

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)
	if err != nil {
		i.logger.Error(err)
		return err
	}

	i.logger.Info(string(output))

	return nil
}

// RestorePipe returns a reader that streams the single file stored in a restic snapshot.
func (i *Restic) RestorePipe(ctx context.Context, snapshotID string) (io.ReadCloser, string, func() error, error) {
	filename, filepath, err := i.getFilenameForRestorePipe(ctx, snapshotID)
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, err
	}

	cmd, err := i.restoreCmd(ctx, snapshotID, filepath, true)
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, err
	}

	reader, err := cmd.StdoutPipe()
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, fmt.Errorf("open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		i.logger.Error(err)
		return nil, "", nil, fmt.Errorf("start restic dump: %w", err)
	}

	return reader, filename, cmd.Wait, nil
}

// IsSupportedRestorePipe reports whether piped restores are supported.
func (i *Restic) IsSupportedRestorePipe(ctx context.Context) bool {
	return true
}
