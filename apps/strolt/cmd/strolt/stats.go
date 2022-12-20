package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/apps/strolt/internal/util"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	statsCmd.Flags().String("service", "", "Service name")
	statsCmd.Flags().StringP("task", "t", "", "Task name")
	statsCmd.Flags().StringP("destination", "d", "", "Destination name")
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print stats information",
	Long:  `All software has versions. This is Strolt's`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		serviceName, err := getServiceName(cmd)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}
		fmt.Printf("selected service: %s", serviceName) //nolint:forbidigo

		taskName, err := getTaskName(cmd, serviceName)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}
		fmt.Println("selected task:", taskName) //nolint:forbidigo

		destinationName, err := getDestinationName(cmd, serviceName, taskName)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}
		fmt.Println("selected destination:", destinationName) //nolint:forbidigo

		t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeSnapshots)

		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}

		defer t.Close()

		var s = &spinner.Spinner{}
		if !isJSONFlag {
			s = util.NewSpinner()
			s.Start()
		}
		stats, err := t.GetStats(destinationName)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}

		if !isJSONFlag {
			s.Stop()
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
		os.Exit(1)
	}

	fmt.Println(string(j)) //nolint:forbidigo
}

func printStatsText(stats interfaces.FormattedStats) {
	fmt.Printf("Total size: %s\n", stats.TotalSizeFormatted)   //nolint:forbidigo
	fmt.Printf("Total file count: %d\n", stats.TotalFileCount) //nolint:forbidigo
	fmt.Printf("Snapshots count: %d\n", stats.SnapshotsCount)  //nolint:forbidigo
}
