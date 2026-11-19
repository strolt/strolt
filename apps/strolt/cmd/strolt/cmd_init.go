package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Aliases: nil, SuggestFor: nil, GroupID: "", Long: "", Example: "", ValidArgs: nil, ValidArgsFunction: nil, Args: nil, ArgAliases: nil, BashCompletionFunction: "", Deprecated: "", Annotations: nil, Version: "", PersistentPreRun: nil, PersistentPreRunE: nil, PreRun: nil, PreRunE: nil, RunE: nil, PostRun: nil, PostRunE: nil, PersistentPostRun: nil, PersistentPostRunE: nil, FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false}, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false, DisableNoDescFlag: false, DisableDescriptions: false, HiddenDefaultCmd: false, DefaultShellCompDirective: nil}, TraverseChildren: false, Hidden: false, SilenceErrors: false, SilenceUsage: false, DisableFlagParsing: false, DisableAutoGenTag: false, DisableFlagsInUseLine: false, DisableSuggestions: false, SuggestionsMinimumDistance: 0,
	Use:   "init",
	Short: "Initialize destinations",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig(cmd.Context())
		c := config.Get()

		for serviceName, service := range c.Services {
			for taskName, task := range service {
				for destinationName, destination := range task.Destinations {
					log := logger.New().WithField("destination", destinationName)
					d, err := dmanager.GetDestinationDriver(cmd.Context(), destinationName, destination.Driver, serviceName, taskName, destination.Config, destination.Env)
					if err != nil {
						log.Error(err)
					} else {
						if err := d.Init(cmd.Context()); err != nil {
							log.Error(err)
						} else {
							log.Info("initialized")
						}
					}
				}
			}
		}
	},
}
