package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	backupCmd.Flags().String("service", "", "service name")
	backupCmd.Flags().String("task", "", "task name")
	backupCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup",
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

		if !isSkipConfirmation && !prompt.AskIsConfirm() {
			return
		}

		t, err := task.New(prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeBackup)
		if err != nil {
			log.Fatal(err)
		}
		defer t.Close()

		if err := t.Backup(); err != nil {
			log.Fatal(err)
		}
	},
}
