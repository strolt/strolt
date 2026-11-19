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
	backupCmd.Flags().String("service", "", "service name")
	backupCmd.Flags().String("task", "", "task name")
	backupCmd.Flags().BoolVar(&isSkipConfirmation, "y", false, "skip confirmation")

	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Aliases:                    nil,
	SuggestFor:                 nil,
	GroupID:                    "",
	Long:                       "",
	Example:                    "",
	ValidArgs:                  nil,
	ValidArgsFunction:          nil,
	Args:                       nil,
	ArgAliases:                 nil,
	BashCompletionFunction:     "",
	Deprecated:                 "",
	Annotations:                nil,
	Version:                    "",
	PersistentPreRun:           nil,
	PersistentPreRunE:          nil,
	PreRun:                     nil,
	PreRunE:                    nil,
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{UnknownFlags: false},
	CompletionOptions:          cobra.CompletionOptions{DisableDefaultCmd: false, DisableNoDescFlag: false, DisableDescriptions: false, HiddenDefaultCmd: false, DefaultShellCompDirective: nil},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
	Use:                        "backup",
	Short:                      "Create backup",
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
			t, err := task.New(cmd.Context(), prompt.ServiceName, prompt.TaskName, sctxt.TManual, sctxt.OpTypeBackup)
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			defer func() {
				if err := t.Close(); err != nil {
					resultErr = errors.Join(resultErr, fmt.Errorf("close task: %w", err))
				}
			}()

			if err := t.Backup(); err != nil {
				return fmt.Errorf("backup task: %w", err)
			}
			return nil
		}(); err != nil {
			log.Fatal(err)
		}
	},
}
