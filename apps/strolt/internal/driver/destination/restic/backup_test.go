package restic

import (
	"testing"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBackupSummary(t *testing.T) {
	summaryLine := `{"message_type":"summary","files_new":1,"files_changed":0,"files_unmodified":0,` +
		`"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":1,"tree_blobs":1,` +
		`"data_added":44040192,"total_files_processed":1,"total_bytes_processed":44040192,` +
		`"total_duration":1.23,"snapshot_id":"a1b2c3d4"}`

	expected := sctxt.BackupOutput{
		FilesNew:            1,
		TotalFilesProcessed: 1,
		TotalBytesProcessed: 44040192,
		SnapshotID:          "a1b2c3d4",
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
