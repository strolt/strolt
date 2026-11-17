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
	pruneCmd.Flags().String("service", "", "service name")
	pruneCmd.Flags().String("task", "", "task name")
	pruneCmd.Flags().String("destination", "", "destination name")
	pruneCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up snapshots",
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
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypePrune)
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
					return fmt.Errorf("confirm prune: %w", err)
				}
				if !confirmed {
					return nil
				}
			}

			if _, err := t.Prune(prompt.DestinationName, false); err != nil {
				return fmt.Errorf("prune destination: %w", err)
			}
			return nil
		}(); err != nil {
			log.Fatal(err)
		}
	},
}
