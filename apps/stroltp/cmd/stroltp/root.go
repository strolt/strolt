// Package cmd implements the stroltp command-line interface.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
	"github.com/strolt/strolt/apps/stroltp/internal/api"
	"github.com/strolt/strolt/apps/stroltp/internal/config"
	"github.com/strolt/strolt/apps/stroltp/internal/env"
	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/strolt"
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
	Use:   "stroltp",
	Short: "Strolt Proxy - reverse proxy and load balancer for Strolt instances",
	Long: `Strolt Proxy (stroltp) provides reverse proxy and load balancing capabilities for Strolt instances.
It enables centralized access management, request routing, and high availability
for your distributed backup infrastructure.`,
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

		{ // Manager
			wg.Add(1)
			go func() {
				// manager.Init().Watch(ctx, cancel)
				instances := make([]strolt.ManagerInstanceInit, 0, len(config.Get().Strolt.Instances))
				for instanceName, instance := range config.Get().Strolt.Instances {
					instances = append(instances, strolt.ManagerInstanceInit{
						Name:     instanceName,
						URL:      instance.URL,
						Username: instance.Username,
						Password: instance.Password, // pragma: allowlist secret
					})
				}
				strolt.ManagerInit(ctx, cancel, instances)
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

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.New().Fatal(err)
	}
}
