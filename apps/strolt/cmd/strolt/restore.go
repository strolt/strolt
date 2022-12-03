package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/apps/strolt/internal/util"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	restoreCmd.Flags().String("service", "", "Service name")
	restoreCmd.Flags().StringP("task", "t", "", "Task name")
	restoreCmd.Flags().StringP("destination", "d", "", "Destination name")
	restoreCmd.Flags().StringP("snapshot", "s", "", "Snapshot")
	restoreCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(restoreCmd)
}

func restoreGetSnapshotName(cmd *cobra.Command, serviceName, taskName string, destinationName string) (string, error) {
	s := util.NewSpinner()
	s.Start()

	t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeRestore)
	if err != nil {
		return "", err
	}
	defer t.Close()

	snapshotList, err := t.GetSnapshotList(destinationName)

	s.Stop()

	if err != nil {
		return "", err
	}

	snapshotName, err := cmd.Flags().GetString("snapshot")
	if err != nil {
		return "", err
	}

	items := make([]string, len(snapshotList))
	for i, snapshot := range snapshotList {
		items[i] = fmt.Sprintf("%s (%s) [%s]", snapshot.ID, snapshot.Time.Format(time.RFC3339), strings.Join(snapshot.Tags, ", "))
	}

	if snapshotName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "Select snapshot",
			Items:        items,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return "", err
		}

		snapshotName = snapshotList[i].ID
	}

	if !snapshotList.IsAvailable(snapshotName) {
		return "", fmt.Errorf("snapshot not exists")
	}

	return snapshotName, nil
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Retore backup",
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

		snapshotName, err := restoreGetSnapshotName(cmd, serviceName, taskName, destinationName)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("selected snapshot: %s", snapshotName)

		if !isSkipConfirmation && !isConfirm() {
			return
		}

		t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeRestore)
		if err != nil {
			log.Error(err)
			return
		}
		defer t.Close()

		isSourceEmpty, err := t.IsSourceEmpty()
		if err != nil {
			log.Error(err)
			return
		}

		if !isSourceEmpty {
			log.Warn("source is not empty")

			if !isSkipConfirmation && !isConfirm() {
				return
			}
		}

		fmt.Println("fetch data...") //nolint:forbidigo
		if err := t.RestoreDestinationToTemp(destinationName, snapshotName); err != nil {
			log.Error(err)
			return
		}

		fmt.Println("restore data...") //nolint:forbidigo
		if err := t.RestoreTempToSource(); err != nil {
			log.Error(err)
			return
		}
	},
}
