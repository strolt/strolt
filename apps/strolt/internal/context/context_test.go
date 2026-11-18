package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

func TestTempWorkDirsAreOwnedIndependently(t *testing.T) {
	t.Setenv("STROLT_PATH_DATA", t.TempDir())
	env.Scan()
	t.Cleanup(func() { env.Scan() })
	first, err := New(sctxt.TManual, "smoke", "task", sctxt.OpTypeBackup, "")
	require.NoError(t, err)
	second, err := New(sctxt.TManual, "smoke", "task", sctxt.OpTypeBackup, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = first.Close()
		_ = second.Close()
	})
	require.NotEqual(t, first.WorkDir, second.WorkDir)
	require.NoError(t, os.WriteFile(filepath.Join(first.WorkDir, "sentinel"), []byte("owned"), 0o600))

	data, err := json.Marshal(first)
	require.NoError(t, err)
	var borrowed Context
	require.NoError(t, json.Unmarshal(data, &borrowed))
	require.NoError(t, borrowed.Close())
	require.FileExists(t, filepath.Join(first.WorkDir, "sentinel"))
	require.NoError(t, first.Close())
	require.NoDirExists(t, first.WorkDir)
	require.NoError(t, os.WriteFile(filepath.Join(second.WorkDir, "sentinel"), []byte("owned"), 0o600))
	require.FileExists(t, filepath.Join(second.WorkDir, "sentinel"))
	require.NoError(t, second.Close())
}

func TestLocalSourcePathRemainsPersistent(t *testing.T) {
	source := t.TempDir()
	ctx, err := New(sctxt.TManual, "smoke", "task", sctxt.OpTypeBackup, source)
	require.NoError(t, err)
	require.Equal(t, source, ctx.WorkDir)
	require.False(t, ctx.IsWorkDirTmp)
	require.NoError(t, ctx.Close())
	require.DirExists(t, source)
}
