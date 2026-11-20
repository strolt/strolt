package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// Snapshot describes a snapshot entry rendered by the snapshots command.
type Snapshot struct {
	ID      string   `json:"id"`
	ShortID string   `json:"shortId"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
}

//nolint:gochecknoinits
func init() {
	snapshotsCmd.Flags().String("service", "", "service name")
	snapshotsCmd.Flags().String("task", "", "task name")
	snapshotsCmd.Flags().String("destination", "", "destination name")
	rootCmd.AddCommand(snapshotsCmd)
}

var snapshotsCmd = &cobra.Command{
	Aliases: nil, SuggestFor: nil, GroupID: "", Long: "", Example: "", ValidArgs: nil, ValidArgsFunction: nil, Args: nil, ArgAliases: nil, BashCompletionFunction: "", Deprecated: "", Annotations: nil, Version: "", PersistentPreRun: nil, PersistentPreRunE: nil, PreRun: nil, PreRunE: nil, RunE: nil, PostRun: nil, PostRunE: nil, PersistentPostRun: nil, PersistentPostRunE: nil, FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false}, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false, DisableNoDescFlag: false, DisableDescriptions: false, HiddenDefaultCmd: false, DefaultShellCompDirective: nil}, TraverseChildren: false, Hidden: false, SilenceErrors: false, SilenceUsage: false, DisableFlagParsing: false, DisableAutoGenTag: false, DisableFlagsInUseLine: false, DisableSuggestions: false, SuggestionsMinimumDistance: 0,
	Use:   "snapshots",
	Short: "Print snapshots",
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

		var snapshotList task.SnapshotList
		if err := func() (resultErr error) {
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeSnapshots)
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			defer func() {
				if err := t.Close(); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("close task: %w", err))
				}
			}()

			snapshotList, err = t.GetSnapshotList(prompt.DestinationName)
			if err != nil {
				return fmt.Errorf("get snapshot list: %w", err)
			}
			return nil
		}(); err != nil {
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

		tbl.AppendRow([]any{snapshotID, snapshot.Time.Format(time.RFC3339), strings.Join(snapshot.Tags, "\n")})

		if i < len(snapshotList)-1 {
			tbl.AppendSeparator()
		}
	}

	tbl.Render()
}

func printSnapshotsJSON(snapshotList task.SnapshotList) {
	_snapshotList := make([]Snapshot, 0, len(snapshotList))

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
