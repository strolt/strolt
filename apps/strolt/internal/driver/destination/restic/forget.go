package restic

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

// snapshotNotFoundMarker is part of the warning restic prints when a forget
// argument matches no snapshot. restic still exits with status 0 in that case,
// so this warning is the only signal that nothing has been removed.
const snapshotNotFoundMarker = "no matching ID found for prefix"

type forgetOutputListItemRemoveListItem struct {
	ID      string    `json:"id"`
	ShortID string    `json:"short_id"`
	Time    time.Time `json:"time"`
	Tags    []string  `json:"tags"`
	Paths   []string  `json:"paths"`
}

type forgetOutputGroupItem struct {
	Remove []forgetOutputListItemRemoveListItem `json:"remove"`
}

type forgetOutput []forgetOutputGroupItem

// isSnapshotNotFound reports whether restic refused to forget the requested
// snapshot because no snapshot matched the given ID.
func isSnapshotNotFound(output []byte) bool {
	return strings.Contains(string(output), snapshotNotFoundMarker)
}

// Forget removes a single snapshot and prunes the data it referenced. Unlike
// Prune it ignores the configured retention policy, so the snapshot is deleted
// even when the keep rules would preserve it.
func (i *Restic) Forget(snapshotID string) error {
	globalFlags := i.getGlobalFlagsWithLock()

	args := make([]string, 0, len(globalFlags)+3) //nolint:mnd // "forget", the snapshot id and "--prune"
	args = append(args, globalFlags...)
	args = append(args, "forget", snapshotID, "--prune")

	cmd := exec.CommandContext(context.Background(), i.getBin(), args...) //nolint:gosec // restic binary path and flags come from validated config

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)

	i.logger.Debug(string(output))

	if err != nil {
		i.logger.Error(err)
		return err
	}

	if isSnapshotNotFound(output) {
		return fmt.Errorf("snapshot '%s' not found in repository", snapshotID)
	}

	return nil
}

func (o *forgetOutput) getSnapshotList() []interfaces.Snapshot {
	list := []interfaces.Snapshot{}

	for _, group := range *o {
		for _, remove := range group.Remove {
			list = append(list, interfaces.Snapshot{
				ID:      remove.ID,
				ShortID: remove.ShortID,
				Time:    remove.Time,
				Tags:    remove.Tags,
				Paths:   remove.Paths,
			})
		}
	}

	return list
}
