package cmd

import (
	"errors"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	forgetCmd.Flags().String("service", "", "service name")
	forgetCmd.Flags().String("task", "", "task name")
	forgetCmd.Flags().String("destination", "", "destination name")
	forgetCmd.Flags().String("snapshot", "", "snapshot id / name")
	forgetCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(forgetCmd)
}

var forgetCmd = &cobra.Command{
	Use:   "forget",
	Short: "Delete a single snapshot, ignoring the retention policy",
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

		if err := prompt.ScanSnapshotName(); err != nil {
			log.Fatal(err)
		}
		Printf("selected snapshot: %s", prompt.SnapshotName)

		if err := func() (resultErr error) {
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeForget)
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			defer func() {
				if err := t.Close(); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("close task: %w", err))
				}
			}()

			if !isSkipConfirmation {
				confirmed, err := prompt.AskIsConfirm()
				if err != nil {
					return fmt.Errorf("confirm forget: %w", err)
				}
				if !confirmed {
					return nil
				}
			}

			if err := t.Forget(prompt.DestinationName, prompt.SnapshotName); err != nil {
				return fmt.Errorf("forget snapshot: %w", err)
			}
			Printf("snapshot removed: %s", prompt.SnapshotName)
			return nil
		}(); err != nil {
			log.Fatal(err)
		}
	},
}
