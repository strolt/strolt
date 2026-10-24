package pg

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
)

// TestWaitWithStderr verifies that a piped process failure carries the
// subprocess stderr tail, so a restore abort names the actual SQL error
// instead of a bare exit status.
func TestWaitWithStderr(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	t.Run("appends stderr tail to a non-zero exit", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", `echo 'psql:<stdin>:23: ERROR:  relation "t" already exists' >&2; exit 3`)
		stderr := bufcap.New(stderrTailCap)
		cmd.Stderr = stderr
		require.NoError(t, cmd.Start())

		err := waitWithStderr(cmd, stderr)()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit status 3")
		assert.Contains(t, err.Error(), `ERROR:  relation "t" already exists`)
	})

	t.Run("returns nil exit unchanged on success", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo note >&2")
		stderr := bufcap.New(stderrTailCap)
		cmd.Stderr = stderr
		require.NoError(t, cmd.Start())

		require.NoError(t, waitWithStderr(cmd, stderr)())
	})

	t.Run("empty stderr keeps the plain exit error", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 9")
		stderr := bufcap.New(stderrTailCap)
		cmd.Stderr = stderr
		require.NoError(t, cmd.Start())

		err := waitWithStderr(cmd, stderr)()
		require.Error(t, err)
		assert.Equal(t, "exit status 9", err.Error())
	})

	t.Run("tail is bounded", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "yes line | head -n 5000 >&2; exit 2")
		stderr := bufcap.New(stderrTailCap)
		cmd.Stderr = stderr
		require.NoError(t, cmd.Start())

		err := waitWithStderr(cmd, stderr)()
		require.Error(t, err)

		msg := err.Error()
		assert.LessOrEqual(t, len(msg), len("exit status 2: ")+stderrTailCap)
		assert.True(t, strings.HasSuffix(strings.TrimSpace(msg), "line"), "the newest stderr must be kept")
	})
}
