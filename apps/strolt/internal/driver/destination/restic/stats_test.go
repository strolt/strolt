package restic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStats(t *testing.T) {
	expected := resticStats{
		TotalSize:      8,
		TotalFileCount: 2,
		SnapshotsCount: 2,
	}

	t.Run("json object preceded by a progress line", func(t *testing.T) {
		// restic 0.19 prints a scan-progress line to stdout before the JSON.
		output := []byte("[0:00] 100.00%  2 / 2 snapshots, 1 files, 4 B\n" +
			`{"total_size":8,"total_file_count":2,"snapshots_count":2}` + "\n")

		got, err := parseStats(output)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("json object only, without trailing newline", func(t *testing.T) {
		got, err := parseStats([]byte(`{"total_size":8,"total_file_count":2,"snapshots_count":2}`))
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("missing json object returns an error", func(t *testing.T) {
		_, err := parseStats([]byte("[0:00] 100.00%  0 / 0 snapshots\n"))
		require.Error(t, err)
	})

	t.Run("malformed json object returns an error", func(t *testing.T) {
		_, err := parseStats([]byte("{not json}"))
		require.Error(t, err)
	})
}
