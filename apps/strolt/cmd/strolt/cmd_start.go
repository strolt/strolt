package cmd

import (
	ctx "context"
	"os"
	"os/signal"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/api"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/schedule"
	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
	"github.com/strolt/strolt/shared"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(startpCmd)
}

var startpCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daemon",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		ctx, cancel := ctx.WithCancel(ctx.Background())
		log := logger.New()

		// GC work directories left over from crashed runs. This lives in the
		// daemon startup (not in every CLI invocation) so concurrent strolt
		// processes sharing the data directory do not wipe each other's
		// in-flight work directories.
		if err := dir.RemoveTempDirectories(); err != nil {
			log.Error(err)
		}

		wg := sync.WaitGroup{}
		c := make(chan os.Signal, 1)
		defer close(c)
		signal.Notify(c, os.Interrupt)

		isConfigChanged := false

		{
			// Api server
			wg.Go(func() {
				api.New().Run(ctx, cancel)
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

		{
			// Watch system exit code
			go func() {
				oscall := <-c
				log.Debugf("system call: %+v", oscall)
				cancel()
			}()
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
