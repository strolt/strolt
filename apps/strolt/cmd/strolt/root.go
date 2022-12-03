package cmd

import (
	ctx "context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/api"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/schedule"
	"github.com/strolt/strolt/apps/strolt/internal/util/dir"

	"github.com/spf13/cobra"
)

const (
	configPath = "./config.yml"
)

var (
	isJSONFlag         = false
	isSkipConfirmation = false
	configPathFlag     = ""
)

func initConfig() {
	if isJSONFlag {
		logger.SetLogFormat(logger.LogFormatJSON)
	}

	if configPathFlag == "" {
		configPathFlag = configPath
	}

	log := logger.New()

	err := config.Load(configPathFlag)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

//nolint:gochecknoinits
func init() {
	env.Scan()
	rootCmd.PersistentFlags().BoolVar(&isJSONFlag, "json", false, "set output mode to JSON")
	// cobra.OnInitialize(initConfig)
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	// rootCmd.PersistentFlags().StringVarP(&projectBase, "projectbase", "b", "", "base project directory eg. github.com/spf13/")
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "Author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project (can provide `licensetext` in config)")
	// rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("projectbase", rootCmd.PersistentFlags().Lookup("projectbase"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")
	rootCmd.PersistentFlags().StringVarP(&configPathFlag, "config", "c", "", fmt.Sprintf("config file (default is %s)", configPath))
}

var rootCmd = &cobra.Command{
	Use:   "strolt",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		parseCliFlags()

		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
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

func Execute() {
	if err := dir.RemoveTempDirectories(); err != nil {
		fmt.Println(err) //nolint:forbidigo
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err) //nolint:forbidigo
		os.Exit(1)
	}

	if err := dir.RemoveTempDirectories(); err != nil {
		fmt.Println(err) //nolint:forbidigo
	}
}
