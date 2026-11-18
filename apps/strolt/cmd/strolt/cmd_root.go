package cmd

import (
	"context"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

const configPath = "./config.yml"

var (
	isJSONFlag         = false
	isSkipConfirmation = false
	configPathFlag     = ""
)

func initConfig(processContext context.Context) {
	if isJSONFlag {
		logger.SetLogFormat(logger.LogFormatJSON)
	}
	if configPathFlag == "" {
		configPathFlag = configPath
	}
	log := logger.New()
	if err := config.Load(processContext, configPathFlag); err != nil {
		log.Fatal(err)
	}
}

//nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().BoolVar(&isJSONFlag, "json", false, "set output mode to JSON")
	rootCmd.PersistentFlags().StringVarP(&configPathFlag, "config", "c", "", fmt.Sprintf("path to config file (default is %s)", configPath))
}

var rootCmd = &cobra.Command{
	Use:   "strolt",
	Short: "Backup and restore tool with support for multiple sources and destinations",
	Long: `strolt is a program for backup and restore with
				support for various sources (filesystem, databases), notifications and API.
				Source code is available at https://github.com/strolt/strolt`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		env.Scan()
		metrics.Init()
		scanCliFlags()
	},
}

// ExecuteContext runs the root command under the supplied cancellation context.
func ExecuteContext(ctx context.Context) error {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}
