package restic

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

// Prune runs restic forget with the configured retention policy and prunes removed data.
func (i *Restic) Prune(ctx context.Context, isDryRun bool) ([]interfaces.Snapshot, error) {
	var args []string

	if isDryRun {
		args = append(args, i.getGlobalFlags()...)
	} else {
		args = append(args, i.getGlobalFlagsWithLock()...)
	}

	args = append(args, "forget", "--group-by=", "--prune")
	args = append(args, i.getKeepFlags()...)

	if isDryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(gocontext.Background(), i.getBin(), args...) //nolint:gosec // restic binary path and flags come from validated config

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)
	if err != nil {
		i.logger.Error(err)
		return []interfaces.Snapshot{}, err
	}

	outputList := strings.Split(string(output), "\n")

	if len(outputList) == 0 || len(outputList[0]) == 0 {
		return []interfaces.Snapshot{}, nil
	}

	var resticForgetOutput forgetOutput
	if err := json.Unmarshal([]byte(outputList[0]), &resticForgetOutput); err != nil {
		return []interfaces.Snapshot{}, fmt.Errorf("unmarshal forget output: %w", err)
	}

	return resticForgetOutput.getSnapshotList(), nil
}
