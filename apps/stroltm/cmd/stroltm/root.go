package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
	"github.com/strolt/strolt/apps/stroltm/internal/api"
	"github.com/strolt/strolt/apps/stroltm/internal/config"
	"github.com/strolt/strolt/apps/stroltm/internal/env"
	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/strolt"
	"github.com/strolt/strolt/shared/sdk/stroltp"
)

const (
	configPath = "./config.yml"
)

var (
	isJSONFlag     = false
	configPathFlag = ""
)

func initConfig() {
	if isJSONFlag {
		logger.SetLogFormat(logger.LogFormatJSON)
	}

	if configPathFlag == "" {
		configPathFlag = configPath
	}

	if err := config.Load(configPathFlag); err != nil {
		logger.New().Fatal(err)
	}
}

//nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().BoolVar(&isJSONFlag, "json", false, "set output mode to JSON")
	rootCmd.PersistentFlags().StringVarP(&configPathFlag, "config", "c", "", fmt.Sprintf("config file (default is %s)", configPath))
}

var rootCmd = &cobra.Command{
	Use:   "stroltm",
	Short: "Strolt Manager - centralized management for Strolt and Stroltp instances",
	Long: `Strolt Manager (stroltm) is a centralized management tool for Strolt and Stroltp instances.
It provides a unified API and web interface for managing multiple backup instances,
monitoring their status, and coordinating backup operations across your infrastructure.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		env.Scan()

		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.New()

		log.Info(config.Get())

		ctx, cancel := context.WithCancel(context.Background())

		wg := sync.WaitGroup{}

		c := make(chan os.Signal, 1)
		defer close(c)

		signal.Notify(c, os.Interrupt)

		{ // Api server
			wg.Add(1)
			go func() {
				api.New().Run(ctx, cancel)
				wg.Done()
			}()
		}

		{ // Strolt Manager
			wg.Add(1)
			go func() {
				// manager.Init().Watch(ctx, cancel)
				instances := []strolt.ManagerInstanceInit{}
				for instanceName, instance := range config.Get().Strolt.Instances {
					instances = append(instances, strolt.ManagerInstanceInit{
						Name:     instanceName,
						URL:      instance.URL,
						Username: instance.Username,
						Password: instance.Password, //pragma: allowlist secret
					})
				}
				strolt.ManagerInit(ctx, cancel, instances)
				wg.Done()
			}()
		}

		{ // Stroltp Manager
			wg.Add(1)
			go func() {
				// manager.Init().Watch(ctx, cancel)
				instances := []stroltp.ManagerInstanceInit{}
				for instanceName, instance := range config.Get().Stroltp.Instances {
					instances = append(instances, stroltp.ManagerInstanceInit{
						Name:     instanceName,
						URL:      instance.URL,
						Username: instance.Username,
						Password: instance.Password, //pragma: allowlist secret
					})
				}
				stroltp.ManagerInit(ctx, cancel, instances)
				wg.Done()
			}()
		}

		// Watch system exit code
		go func() {
			oscall := <-c
			log.Debugf("system call: %+v", oscall)
			cancel()
		}()

		wg.Wait()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.New().Fatal(err)
	}
}
