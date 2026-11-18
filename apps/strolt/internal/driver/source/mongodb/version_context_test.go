package mongodb

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

func TestBinaryVersionFailureStopsAtEachMongoBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	binaries := []struct {
		name   string
		output string
	}{
		{"mongodump", "mongodump version: 100.13.0\n"},
		{"mongorestore", "mongorestore version: 100.13.0\n"},
	}
	for failed := range binaries {
		t.Run(binaries[failed].name, func(t *testing.T) {
			root := t.TempDir()
			marker := filepath.Join(root, "later-invoked")
			paths := make([]string, len(binaries))
			for index, binary := range binaries {
				paths[index] = filepath.Join(root, binary.name)
				contents := "#!/bin/sh\nprintf '%s' " + shellQuote(binaries[index].output) + "\n"
				if index == failed {
					contents = "#!/bin/sh\nprintf 'probe diagnostic\\n' >&2\nexit 17\n"
				} else if index > failed {
					contents += "touch " + shellQuote(marker) + "\n"
				}
				require.NoError(t, os.WriteFile(paths[index], []byte(contents), 0o700))
			}
			driver := &MongoDB{config: Config{BinPathMongoDump: paths[0], BinPathMongoRestore: paths[1]}}
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

func TestBinaryVersionReturnsParsedMongoVersions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	root := t.TempDir()
	outputs := []string{"mongodump version: 100.13.0\n", "mongorestore version: 100.13.0\n"}
	paths := make([]string, len(outputs))
	for index, output := range outputs {
		paths[index] = filepath.Join(root, fmt.Sprintf("binary-%d", index))
		require.NoError(t, os.WriteFile(paths[index], []byte("#!/bin/sh\nprintf '%s' "+shellQuote(output)+"\n"), 0o700))
	}
	driver := &MongoDB{config: Config{BinPathMongoDump: paths[0], BinPathMongoRestore: paths[1]}}
	versions, err := driver.BinaryVersion(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"100.13.0", "100.13.0"}, []string{versions[0].Version, versions[1].Version})
}

func TestBinaryVersionCanceledAndMissingMongoProbe(t *testing.T) {
	driver := &MongoDB{config: Config{BinPathMongoDump: filepath.Join(t.TempDir(), "missing")}}
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

func TestBinaryVersionCancelsRunningMongoProbe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	root := t.TempDir()
	ready := filepath.Join(root, "ready")
	marker := filepath.Join(root, "later-invoked")
	first := filepath.Join(root, "mongodump")
	second := filepath.Join(root, "mongorestore")
	require.NoError(t, os.WriteFile(first, []byte("#!/bin/sh\nprintf ready > "+shellQuote(ready)+"\nwhile :; do /bin/sleep 1; done\n"), 0o700))
	require.NoError(t, os.WriteFile(second, []byte("#!/bin/sh\ntouch "+shellQuote(marker)+"\nprintf 'mongorestore version: 100.13.0\\n'\n"), 0o700))
	driver := &MongoDB{config: Config{BinPathMongoDump: first, BinPathMongoRestore: second}}
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
