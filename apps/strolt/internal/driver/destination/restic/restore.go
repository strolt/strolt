package restic

import (
	"fmt"
	"io"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
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

// sizedReadCloser exposes the stream size reported by `restic ls` so the
// restore progress can show a percentage without changing the driver
// interface. Drivers without size knowledge stay unchanged.
type sizedReadCloser struct {
	io.ReadCloser

	size    uint64
	hasSize bool
}

// Size returns the total size of the streamed dump when restic reported it.
func (r *sizedReadCloser) Size() (uint64, bool) { return r.size, r.hasSize }

// RestorePipe returns a reader that streams the single file stored in a restic snapshot.
func (i *Restic) RestorePipe(ctx context.Context, snapshotID string) (io.ReadCloser, string, func() error, error) {
	filename, filepath, size, hasSize, err := i.getFilenameForRestorePipe(ctx, snapshotID)
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, err
	}

	cmd, err := i.restoreCmd(ctx, snapshotID, filepath, true)
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, err
	}

	stderr := bufcap.New(stderrTailCap)
	cmd.Stderr = stderr

	reader, err := cmd.StdoutPipe()
	if err != nil {
		i.logger.Error(err)
		return nil, "", nil, fmt.Errorf("open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		i.logger.Error(err)
		return nil, "", nil, fmt.Errorf("start restic dump: %w", err)
	}

	return &sizedReadCloser{ReadCloser: reader, size: size, hasSize: hasSize}, filename, waitWithStderr(cmd, stderr), nil
}

// IsSupportedRestorePipe reports whether piped restores are supported.
func (i *Restic) IsSupportedRestorePipe(ctx context.Context) bool {
	return true
}
