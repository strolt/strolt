package cmd

import (
	ctx "context"
	"os"
	"os/signal"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/api"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/schedule"
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

		wg := sync.WaitGroup{}
		c := make(chan os.Signal, 1)
		defer close(c)
		signal.Notify(c, os.Interrupt)

		{
			// Api server
			wg.Add(1)
			go func() {
				api.New().Run(ctx, cancel)
				wg.Done()
			}()
		}

		{
			// Watch config
			wg.Add(1)
			go func() {
				config.WatchConfigChanges(ctx, cancel)
				wg.Done()
			}()
		}

		{
			// Schedule manager
			wg.Add(1)
			go func() {
				schedule.Run(ctx)
				wg.Done()
			}()
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
	},
}
