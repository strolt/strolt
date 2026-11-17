package restic

import (
	gocontext "context"
	"fmt"

	"github.com/strolt/strolt/shared/process"
)

// Init initializes the restic repository.
func (i *Restic) Init(processContext gocontext.Context) error {
	globalFlags := i.getGlobalFlags()

	args := make([]string, 0, len(globalFlags)+1)
	args = append(args, globalFlags...)
	args = append(args, "init")

	cmd := process.CommandContext(processContext, i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := cmd.CombinedOutput()
	i.logger.Debug(string(output))

	if err != nil {
		return fmt.Errorf("restic init: %w", err)
	}

	return nil
}
