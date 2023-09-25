package restic

import (
	"io"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

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
		return nil, "", nil, err
	}

	if err := cmd.Start(); err != nil {
		i.logger.Error(err)
		return nil, "", nil, err
	}

	return reader, filename, cmd.Wait, nil
}

func (i *Restic) IsSupportedRestorePipe(ctx context.Context) bool {
	return true
}
