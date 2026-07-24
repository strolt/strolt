package restic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

type resticStats struct {
	TotalSize      uint64 `json:"total_size"`
	TotalFileCount uint64 `json:"total_file_count"`
	SnapshotsCount int    `json:"snapshots_count"`
}

// parseStats extracts the stats object from restic's --json stats output.
// restic prints a scan-progress line (e.g. "[0:00] 100.00%  2 / 2 snapshots")
// to stdout before the JSON object, so the whole output is not valid JSON on
// its own. Scan from the end for the line holding the JSON object and ignore
// the progress line.
func parseStats(output []byte) (resticStats, error) {
	for _, line := range slices.Backward(strings.Split(string(output), "\n")) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var stats resticStats
		if err := json.Unmarshal([]byte(line), &stats); err != nil {
			return resticStats{}, fmt.Errorf("unmarshal stats output: %w", err)
		}

		return stats, nil
	}

	return resticStats{}, errors.New("restic stats output not found")
}

// Stats returns repository statistics reported by restic.
func (i *Restic) Stats() (interfaces.Stats, error) {
	cmd := exec.CommandContext(context.Background(), i.getBin(), "--json", "stats") //nolint:gosec // restic binary path comes from validated config

	env, err := i.getEnv()
	if err != nil {
		return interfaces.Stats{}, err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)

	i.logger.Debug(string(output))

	if err != nil {
		i.logger.Error(err)
		return interfaces.Stats{}, err
	}

	parsed, err := parseStats(output)
	if err != nil {
		i.logger.Error(err)
		return interfaces.Stats{}, err
	}

	stats := interfaces.Stats{
		TotalSize:      parsed.TotalSize,
		TotalFileCount: parsed.TotalFileCount,
		SnapshotsCount: parsed.SnapshotsCount,
	}

	return stats, nil
}
