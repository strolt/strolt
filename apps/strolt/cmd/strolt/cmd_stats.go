package cmd

import (
	"encoding/json"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	statsCmd.Flags().String("service", "", "Service name")
	statsCmd.Flags().String("task", "", "Task name")
	statsCmd.Flags().String("destination", "", "Destination name")
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print destination stats",
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

		t, err := task.New(prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeSnapshots)
		if err != nil {
			log.Fatal(err)
		}

		defer t.Close()

		stats, err := t.GetStats(prompt.DestinationName)
		if err != nil {
			log.Fatal(err)
		}

		if isJSONFlag {
			printStatsJSON(stats)
		} else {
			printStatsText(stats)
		}
	},
}

func printStatsJSON(stats interfaces.FormattedStats) {
	j, err := json.Marshal(stats)
	if err != nil {
		logger.New().Fatal(err)
	}

	Print(string(j))
}

func printStatsText(stats interfaces.FormattedStats) {
	log := logger.New()

	log.Infof("total size: %s\n", stats.TotalSizeFormatted)
	log.Infof("total file count: %d\n", stats.TotalFileCount)
	log.Infof("snapshots count: %d\n", stats.SnapshotsCount)
}
