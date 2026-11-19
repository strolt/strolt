package cmd

import (
	"errors"
	"fmt"

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
	Aliases: nil, SuggestFor: nil, GroupID: "", Long: "", Example: "", ValidArgs: nil, ValidArgsFunction: nil, Args: nil, ArgAliases: nil, BashCompletionFunction: "", Deprecated: "", Annotations: nil, Version: "", PersistentPreRun: nil, PersistentPreRunE: nil, PreRun: nil, PreRunE: nil, RunE: nil, PostRun: nil, PostRunE: nil, PersistentPostRun: nil, PersistentPostRunE: nil, FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false}, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false, DisableNoDescFlag: false, DisableDescriptions: false, HiddenDefaultCmd: false, DefaultShellCompDirective: nil}, TraverseChildren: false, Hidden: false, SilenceErrors: false, SilenceUsage: false, DisableFlagParsing: false, DisableAutoGenTag: false, DisableFlagsInUseLine: false, DisableSuggestions: false, SuggestionsMinimumDistance: 0,
	Use:   "restore",
	Short: "Restore snapshot",
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

		if err := prompt.ScanSnapshotName(); err != nil {
			log.Fatal(err)
		}
		Printf("selected snapshot: %s", prompt.SnapshotName)

		if !isSkipConfirmation {
			confirmed, err := prompt.AskIsConfirm()
			if err != nil {
				log.Fatal(err)
			}
			if !confirmed {
				return
			}
		}

		if err := func() (resultErr error) {
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeRestore)
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			defer func() {
				if err := t.Close(); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("close task: %w", err))
				}
			}()

			isSourceEmpty, err := t.IsSourceEmpty()
			if err != nil {
				return fmt.Errorf("check source emptiness: %w", err)
			}

			if !isSourceEmpty {
				log.Warn("source is not empty")

				if !isSkipConfirmation {
					confirmed, err := prompt.AskIsConfirm()
					if err != nil {
						return fmt.Errorf("confirm restore: %w", err)
					}
					if !confirmed {
						return nil
					}
				}
			}

			if err := t.Restore(prompt.DestinationName, prompt.SnapshotName); err != nil {
				return fmt.Errorf("restore task: %w", err)
			}
			return nil
		}(); err != nil {
			log.Fatal(err)
		}
	},
}
