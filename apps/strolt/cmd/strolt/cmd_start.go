package cmd

import (
	ctx "context"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/api"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/schedule"
	"github.com/strolt/strolt/shared"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(startpCmd)
}

var startpCmd = &cobra.Command{
	Aliases: nil, SuggestFor: nil, GroupID: "", Long: "", Example: "", ValidArgs: nil, ValidArgsFunction: nil, Args: nil, ArgAliases: nil, BashCompletionFunction: "", Deprecated: "", Annotations: nil, Version: "", PersistentPreRun: nil, PersistentPreRunE: nil, PreRun: nil, PreRunE: nil, RunE: nil, PostRun: nil, PostRunE: nil, PersistentPostRun: nil, PersistentPostRunE: nil, FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: false}, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false, DisableNoDescFlag: false, DisableDescriptions: false, HiddenDefaultCmd: false, DefaultShellCompDirective: nil}, TraverseChildren: false, Hidden: false, SilenceErrors: false, SilenceUsage: false, DisableFlagParsing: false, DisableAutoGenTag: false, DisableFlagsInUseLine: false, DisableSuggestions: false, SuggestionsMinimumDistance: 0,
	Use:   "start",
	Short: "Start daemon",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig(cmd.Context())
		ctx, cancel := ctx.WithCancel(cmd.Context())
		defer cancel()
		log := logger.New()

		wg := sync.WaitGroup{}

		isConfigChanged := false

		{
			// Api server
			wg.Go(func() {
				api.New(ctx).Run(ctx, cancel)
			})
		}

		{
			// Watch config
			wg.Go(func() {
				config.WatchConfigChanges(ctx, func() {
					isConfigChanged = true
					cancel()
				})
			})
		}

		{
			// Schedule manager
			wg.Go(func() {
				schedule.Run(ctx)
			})
		}

		wg.Wait()

		if isConfigChanged {
			log.Debugf("restart self...")
			if err := shared.RestartSelf(); err != nil {
				log.Error(err)
			}
		}
	},
}
