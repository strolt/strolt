package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	restoreCmd.Flags().String("service", "", "service name")
	restoreCmd.Flags().String("task", "", "task name")
	restoreCmd.Flags().String("destination", "", "destination name")
	restoreCmd.Flags().String("snapshot", "", "snapshot id / name")
	restoreCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
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

		if !isSkipConfirmation && !prompt.AskIsConfirm() {
			return
		}

		t, err := task.New(prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeRestore)
		if err != nil {
			log.Fatal(err)
		}
		defer t.Close()

		isSourceEmpty, err := t.IsSourceEmpty()
		if err != nil {
			log.Fatal(err)
		}

		if !isSourceEmpty {
			log.Warn("source is not empty")

			if !isSkipConfirmation && !prompt.AskIsConfirm() {
				return
			}
		}

		if err := t.Restore(prompt.DestinationName, prompt.SnapshotName); err != nil {
			log.Fatal(err)
		}
	},
}
