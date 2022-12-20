package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	pruneCmd.Flags().String("service", "", "Service name")
	pruneCmd.Flags().StringP("task", "t", "", "Task name")
	pruneCmd.Flags().StringP("destination", "d", "", "Destination name")
	pruneCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "prune",
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

		destinationName, err := getDestinationName(cmd, serviceName, taskName)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("selected destination: %s", destinationName)

		t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeRestore)
		if err != nil {
			log.Error(err)
			return
		}
		defer t.Close()

		if !isSkipConfirmation && !isConfirm() {
			return
		}

		log.Info("prune...")

		_, err = t.Prune(destinationName, false)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info("success")
	},
}
