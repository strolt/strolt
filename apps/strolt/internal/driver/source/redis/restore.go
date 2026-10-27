package redis

import (
	"errors"
	"io"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

// Restore is not supported: an RDB file can only be loaded by placing it in
// the server's data directory and restarting the instance, which strolt
// cannot do over the Redis protocol.
func (i *Redis) Restore(_ context.Context) error {
	return errors.New("not support restore")
}

// RestorePipe is not supported by the Redis driver.
func (i *Redis) RestorePipe(_ context.Context, _ string) (io.WriteCloser, func() error, error) {
	return nil, func() error { return nil }, errors.New("not support pipe")
}

// IsSupportedRestorePipe reports whether restoring from a stream is supported.
func (i *Redis) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}
