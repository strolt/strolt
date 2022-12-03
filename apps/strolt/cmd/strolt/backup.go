package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	backupCmd.Flags().String("service", "", "Service name")
	backupCmd.Flags().StringP("task", "t", "", "Task name")
	backupCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup",
	Long:  `All software has versions. This is Strolt's`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		log := logger.New()

		serviceName, err := getServiceName(cmd)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("selected service: %s", serviceName)

		taskName, err := getTaskName(cmd, serviceName)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("selected task: %s", taskName)

		if !isSkipConfirmation && !isConfirm() {
			return
		}

		t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeBackup)
		if err != nil {
			log.Error(err)
			return
		}
		defer t.Close()

		log.Info("backup source")
		if err = t.BackupSourceToWorkDir(); err != nil {
			log.Error(err)
			return
		}

		for destinationName := range t.TaskConfig.Destinations {
			log := log.WithField(destinationName, destinationName)
			log.Info("backup destination")
			_, err := t.BackupWorkDirToDestination(destinationName)

			if err != nil {
				log.Error(err)
			}
		}
	},
}
