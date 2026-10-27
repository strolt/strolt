package redis

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	driverContext "github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/shared/logger"
)

// writeScript creates an executable fake redis-cli with the given body. The
// pipe failure paths are subprocess orchestration (EPIPE teardown, stderr
// capture), so they are exercised against a real process, not a mock.
func writeScript(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-redis-cli")
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755))

	return path
}

func newPipeDriver(t *testing.T, body string) *Redis {
	t.Helper()

	r := New()
	r.SetLogger(logger.New())
	require.NoError(t, r.SetConfig(map[string]any{
		"host":               "redis",
		"bin_path_redis_cli": writeScript(t, body),
	}))

	return r
}

// TestBackupPipe_startupFailureCarriesStderr pins the diagnostics contract:
// a redis-cli that fails before producing stdout (connection refused,
// NOAUTH) must fail the pipe with its actual stderr, not a bare EOF.
func TestBackupPipe_startupFailureCarriesStderr(t *testing.T) {
	r := newPipeDriver(t, `echo "Could not connect to Redis at redis:6379: Try again" >&2; exit 1`)

	reader, _, _, err := r.BackupPipe(driverContext.Context{WorkDir: t.TempDir()})
	require.Error(t, err)
	assert.Nil(t, reader)
	assert.Contains(t, err.Error(), "invalid RDB stream")
	assert.Contains(t, err.Error(), "Could not connect to Redis at redis:6379")
}

// TestBackupPipe_badMagicAbortsProducer pins the teardown contract: a
// producer emitting a non-RDB stream must be killed instead of waited on,
// and its stderr must still explain the failure.
func TestBackupPipe_badMagicAbortsProducer(t *testing.T) {
	r := newPipeDriver(t, `printf 'NOTREDIS'; echo "transfer failed" >&2; sleep 30`)

	started := time.Now()
	reader, _, _, err := r.BackupPipe(driverContext.Context{WorkDir: t.TempDir()})
	require.Error(t, err)
	assert.Nil(t, reader)
	assert.Contains(t, err.Error(), `invalid RDB stream: got "NOTRE"`)
	assert.Contains(t, err.Error(), "transfer failed")
	assert.Less(t, time.Since(started), 10*time.Second, "abort must kill the producer, not wait out its sleep")
}
