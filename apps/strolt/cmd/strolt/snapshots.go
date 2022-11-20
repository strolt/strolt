package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/apps/strolt/internal/util"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type Snapshot struct {
	ID      string `json:"id"`
	ShortID string `json:"shortId"`
	Date    string `json:"date"`
	Tags    []string
}

//nolint:gochecknoinits
func init() {
	snapshotsCmd.Flags().String("service", "", "Service name")
	snapshotsCmd.Flags().StringP("task", "t", "", "Task name")
	snapshotsCmd.Flags().StringP("destination", "d", "", "Destination name")
	rootCmd.AddCommand(snapshotsCmd)
}

var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "Print snapshots information",
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
		snapshotList, err := t.GetSnapshotList(destinationName)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			return
		}

		if !isJSONFlag {
			s.Stop()
		}

		if isJSONFlag {
			printSnapshotsJSON(snapshotList)
		} else {
			printSnapshotsTable(snapshotList)
		}
	},
}

func printSnapshotsTable(snapshotList task.SnapshotList) {
	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"ID", "Time", "Tags"})

	for i, snapshot := range snapshotList {
		tbl.AppendRow([]interface{}{snapshot.ID, snapshot.Time.Format(time.RFC3339), strings.Join(snapshot.Tags, "\n")})

		if i < len(snapshotList)-1 {
			tbl.AppendSeparator()
		}
	}

	tbl.Render()
}

func printSnapshotsJSON(snapshotList task.SnapshotList) {
	var _snapshotList = []Snapshot{}

	for _, snapshot := range snapshotList {
		_snapshotList = append(_snapshotList, Snapshot{
			ID:      snapshot.ID,
			ShortID: snapshot.ShortID,
			Date:    snapshot.Time.Format(time.RFC3339),
			Tags:    snapshot.Tags,
		})
	}

	j, err := json.Marshal(_snapshotList)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(string(j)) //nolint:forbidigo
}
