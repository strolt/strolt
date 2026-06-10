package restic

import (
	"encoding/json"
	"testing"
	"time"
)

// A trimmed-down `restic forget --json` response: two groups, one of them
// with nothing to remove (restic emits `"remove": null` in that case).
const forgetOutputFixture = `[
	{
		"remove": [
			{
				"id": "0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b",
				"short_id": "0a1b2c3d",
				"time": "2020-01-01T10:00:00Z",
				"tags": ["trigger=MANUAL"],
				"paths": ["/e2e/input"]
			},
			{
				"id": "ffeeddccbbaaffeeddccbbaaffeeddccbbaaffeeddccbbaaffeeddccbbaaffee",
				"short_id": "ffeeddcc",
				"time": "2020-01-02T09:15:00Z",
				"tags": [],
				"paths": ["/e2e/input"]
			}
		]
	},
	{
		"remove": null
	}
]`

func TestForgetOutputGetSnapshotList(t *testing.T) {
	t.Parallel()

	var output forgetOutput
	if err := json.Unmarshal([]byte(forgetOutputFixture), &output); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	snapshots := output.getSnapshotList()

	if len(snapshots) != 2 {
		t.Fatalf("getSnapshotList() returned %d snapshots, want 2: %+v", len(snapshots), snapshots)
	}

	first := snapshots[0]
	if first.ID != "0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b2c3d4e5f0a1b" ||
		first.ShortID != "0a1b2c3d" ||
		!first.Time.Equal(time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)) ||
		len(first.Tags) != 1 || first.Tags[0] != "trigger=MANUAL" ||
		len(first.Paths) != 1 || first.Paths[0] != "/e2e/input" {
		t.Fatalf("first snapshot parsed incorrectly: %+v", first)
	}

	if snapshots[1].ShortID != "ffeeddcc" {
		t.Fatalf("second snapshot parsed incorrectly: %+v", snapshots[1])
	}
}

func TestForgetOutputGetSnapshotListEmpty(t *testing.T) {
	t.Parallel()

	var output forgetOutput
	if err := json.Unmarshal([]byte(`[{"remove": null}]`), &output); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if snapshots := output.getSnapshotList(); len(snapshots) != 0 {
		t.Fatalf("getSnapshotList() = %+v, want empty", snapshots)
	}
}
