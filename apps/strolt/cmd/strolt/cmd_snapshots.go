package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

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
	snapshotsCmd.Flags().String("service", "", "service name")
	snapshotsCmd.Flags().String("task", "", "task name")
	snapshotsCmd.Flags().String("destination", "", "destination name")
	rootCmd.AddCommand(snapshotsCmd)
}

var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "Print snapshots",
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

		snapshotList, err := t.GetSnapshotList(prompt.DestinationName)
		if err != nil {
			log.Fatal(err)
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
		snapshotID := snapshot.ShortID

		if snapshotID == "" {
			snapshotID = snapshot.ID
		}

		tbl.AppendRow([]interface{}{snapshotID, snapshot.Time.Format(time.RFC3339), strings.Join(snapshot.Tags, "\n")})

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
		logger.New().Fatal(err)
	}

	Print(string(j))
}
