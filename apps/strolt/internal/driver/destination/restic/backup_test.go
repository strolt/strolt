package restic

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
	"github.com/strolt/strolt/shared/process"
)

func TestParseBackupSummary(t *testing.T) {
	summaryLine := `{"message_type":"summary","files_new":1,"files_changed":0,"files_unmodified":0,` +
		`"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":1,"tree_blobs":1,` +
		`"data_added":44040192,"total_files_processed":1,"total_bytes_processed":44040192,` +
		`"total_duration":1.23,"snapshot_id":"a1b2c3d4"}`

	expected := sctxt.BackupOutput{
		FilesNew:            1,
		FilesChanged:        0,
		FilesUnmodified:     0,
		DirsNew:             0,
		DirsChanged:         0,
		DirsUnmodified:      0,
		TotalFilesProcessed: 1,
		TotalBytesProcessed: 44040192,
		SnapshotID:          "a1b2c3d4",
		TotalBytesStreamed:  0,
	}

	t.Run("summary after status lines with trailing newline", func(t *testing.T) {
		output := []byte(`{"message_type":"status","percent_done":0}` + "\n" +
			`{"message_type":"status","percent_done":0.5}` + "\n" +
			summaryLine + "\n")

		got, err := parseBackupSummary(output)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("summary without trailing newline", func(t *testing.T) {
		got, err := parseBackupSummary([]byte(summaryLine))
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("streamed bytes are not populated by restic summary", func(t *testing.T) {
		got, err := parseBackupSummary([]byte(summaryLine))
		require.NoError(t, err)
		assert.Zero(t, got.TotalBytesStreamed)
	})

	t.Run("missing summary returns an error", func(t *testing.T) {
		output := []byte(`{"message_type":"status","percent_done":0}` + "\n" +
			`{"message_type":"status","percent_done":1}` + "\n")

		_, err := parseBackupSummary(output)
		require.Error(t, err)
	})

	t.Run("invalid json returns an error", func(t *testing.T) {
		_, err := parseBackupSummary([]byte("not json"))
		require.Error(t, err)
	})
}

func TestStartCmdReturnsSubprocessFailure(t *testing.T) {
	if os.Getenv("RESTIC_HELPER") == "1" {
		fmt.Fprintln(os.Stderr, "fixture failure")
		os.Exit(7)
	}

	cmd := process.Command(os.Args[0], "-test.run=TestStartCmdReturnsSubprocessFailure")
	cmd.Env = append(os.Environ(), "RESTIC_HELPER=1")

	_, err := startCmd(cmd)
	var exitErr *process.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 7, exitErr.ExitCode())
	assert.Contains(t, err.Error(), "fixture failure")
}

func TestWaitWithStderrRetainsStreamingFailure(t *testing.T) {
	if os.Getenv("RESTIC_HELPER") == "1" {
		fmt.Fprintln(os.Stderr, "stream fixture failure")
		os.Exit(7)
	}

	cmd := process.Command(os.Args[0], "-test.run=TestWaitWithStderrRetainsStreamingFailure")
	cmd.Env = append(os.Environ(), "RESTIC_HELPER=1")
	stderr := bufcap.New(stderrTailCap)
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	err := waitWithStderr(cmd, stderr)()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stream fixture failure")
}

func TestBackupPipeRetainsStreamingFailure(t *testing.T) {
	if os.Getenv("RESTIC_HELPER") == "1" {
		fmt.Fprintln(os.Stderr, "backup fixture failure")
		os.Exit(7)
	}

	cmd := process.Command(os.Args[0], "-test.run=TestBackupPipeRetainsStreamingFailure")
	cmd.Env = append(os.Environ(), "RESTIC_HELPER=1")
	cmd.Stdin = strings.NewReader("")
	cmd.Stdout = io.Discard
	stderr := bufcap.New(stderrTailCap)
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	err := waitWithStderr(cmd, stderr)()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup fixture failure")
}
