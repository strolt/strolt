package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

type resticStats struct {
	TotalSize      uint64 `json:"total_size"`
	TotalFileCount uint64 `json:"total_file_count"`
	SnapshotsCount int    `json:"snapshots_count"`
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

	resticStats := resticStats{}
	if err := json.Unmarshal(output, &resticStats); err != nil {
		i.logger.Error(err)
		return interfaces.Stats{}, fmt.Errorf("unmarshal stats output: %w", err)
	}

	stats := interfaces.Stats{
		TotalSize:      resticStats.TotalSize,
		TotalFileCount: resticStats.TotalFileCount,
		SnapshotsCount: resticStats.SnapshotsCount,
	}

	return stats, nil
}
