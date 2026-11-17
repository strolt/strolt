package cmd

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

var isRemoveAllLocks = false

//nolint:gochecknoinits
func init() {
	unlockCmd.Flags().String("service", "", "service name")
	unlockCmd.Flags().String("task", "", "task name")
	unlockCmd.Flags().String("destination", "", "destination name")
	unlockCmd.Flags().BoolVar(&isRemoveAllLocks, "remove-all", false, "remove all locks, including those held by running operations")

	rootCmd.AddCommand(unlockCmd)
}

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Remove locks from the destination repository",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig(cmd.Context())
		log := logger.New()
		prompt := NewPrompt(cmd)

		if err := prompt.ScanServiceName(); err != nil {
			log.Fatal(err)
		}
		Printf("selected service: %s", prompt.ServiceName)

		if err := prompt.ScanTaskName(); err != nil {
			log.Fatal(err)
		}
		Printf("selected task: %s", prompt.TaskName)

		if err := prompt.ScanDestinationName(); err != nil {
			log.Fatal(err)
		}
		Printf("selected destination: %s", prompt.DestinationName)

		if err := func() (resultErr error) {
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeUnlock)
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			defer func() {
				if err := t.Close(); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("close task: %w", err))
				}
			}()

			if err := t.Unlock(prompt.DestinationName, isRemoveAllLocks); err != nil {
				return fmt.Errorf("unlock destination: %w", err)
			}
			Printf("destination unlocked: %s", prompt.DestinationName)
			return nil
		}(); err != nil {
			log.Fatal(err)
		}
	},
}
