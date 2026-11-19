package pg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/shared/process"
)

func TestBinaryVersionFailureStopsAtEachPostgresBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	binaries := []struct {
		name   string
		output string
	}{
		{"pg_dump", "pg_dump (PostgreSQL) 18.7\n"},
		{"pg_restore", "pg_restore (PostgreSQL) 18.7\n"},
		{"psql", "psql (PostgreSQL) 18.7\n"},
	}
	for failed := range binaries {
		t.Run(binaries[failed].name, func(t *testing.T) {
			root := t.TempDir()
			marker := filepath.Join(root, "later-invoked")
			paths := make([]string, len(binaries))
			for index, binary := range binaries {
				paths[index] = filepath.Join(root, binary.name)
				contents := "#!/bin/sh\nprintf '%s' " + shellQuote(binary.output) + "\n"
				if index == failed {
					contents = "#!/bin/sh\nprintf 'probe diagnostic\\n' >&2\nexit 17\n"
				} else if index > failed {
					contents += "touch " + shellQuote(marker) + "\n"
				}
				require.NoError(t, os.WriteFile(paths[index], []byte(contents), 0o700))
			}
			driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: paths[0], BinPathPgRestore: paths[1], BinPathPsql: paths[2]}}
			versions, err := driver.BinaryVersion(context.Background())
			require.Nil(t, versions)
			var exitError *process.ExitError
			require.ErrorAs(t, err, &exitError)
			require.Equal(t, 17, exitError.ExitCode())
			require.ErrorContains(t, err, "probe diagnostic")
			require.NoFileExists(t, marker)
		})
	}
}

func TestBinaryVersionReturnsParsedPostgresVersions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	root := t.TempDir()
	outputs := []string{"pg_dump (PostgreSQL) 18.7\n", "pg_restore (PostgreSQL) 18.7\n", "psql (PostgreSQL) 18.7\n"}
	paths := make([]string, len(outputs))
	for index, output := range outputs {
		paths[index] = filepath.Join(root, fmt.Sprintf("binary-%d", index))
		require.NoError(t, os.WriteFile(paths[index], []byte("#!/bin/sh\nprintf '%s' "+shellQuote(output)+"\n"), 0o700))
	}
	driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: paths[0], BinPathPgRestore: paths[1], BinPathPsql: paths[2]}}
	versions, err := driver.BinaryVersion(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"18.7\n", "18.7\n", "18.7\n"}, []string{versions[0].Version, versions[1].Version, versions[2].Version})
}

func TestBinaryVersionCanceledAndMissingPostgresProbe(t *testing.T) {
	driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{BinPathPgRestore: "", BinPathPsql: "", Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: filepath.Join(t.TempDir(), "missing")}}
	versions, err := driver.BinaryVersion(context.Background())
	require.Nil(t, versions)
	require.Error(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	versions, err = driver.BinaryVersion(ctx)
	require.Nil(t, versions)
	require.ErrorIs(t, err, context.Canceled)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestBinaryVersionCancelsRunningPostgresProbe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	root := t.TempDir()
	ready := filepath.Join(root, "ready")
	marker := filepath.Join(root, "later-invoked")
	first := filepath.Join(root, "pg_dump")
	second := filepath.Join(root, "pg_restore")
	require.NoError(t, os.WriteFile(first, []byte("#!/bin/sh\nprintf ready > "+shellQuote(ready)+"\nwhile :; do /bin/sleep 1; done\n"), 0o700))
	require.NoError(t, os.WriteFile(second, []byte("#!/bin/sh\ntouch "+shellQuote(marker)+"\nprintf 'pg_restore (PostgreSQL) 18.7\\n'\n"), 0o700))
	driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{BinPathPsql: "", Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: first, BinPathPgRestore: second}}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		versions, err := driver.BinaryVersion(ctx)
		if versions != nil {
			result <- fmt.Errorf("unexpected partial versions: %v", versions)
			return
		}
		result <- err
	}()
	require.Eventually(t, func() bool {
		_, err := os.Stat(ready)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	cancel()
	err := <-result
	require.ErrorIs(t, err, context.Canceled)
	var exitError *process.ExitError
	require.ErrorAs(t, err, &exitError)
	require.NoFileExists(t, marker)
}
