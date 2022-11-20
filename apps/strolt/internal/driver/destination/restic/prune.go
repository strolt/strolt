package restic

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *Restic) Prune(ctx context.Context, isDryRun bool) ([]interfaces.Snapshot, error) {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "forget", "--group-by=", "--prune")
	args = append(args, i.getKeepFlags()...)

	if isDryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command(i.getBin(), args...)

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
		return []interfaces.Snapshot{}, err
	}

	return resticForgetOutput.getSnapshotList(), nil
}
