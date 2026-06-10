package cmd

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/shared/logger"

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

// Execute runs the root command of the strolt CLI.
//
// The shared temp directory is deliberately NOT wiped here: several strolt
// processes can share one data directory (the daemon plus CLI invocations
// inside the same container), and a blanket cleanup on every CLI run deletes
// the work directories of operations still in flight. Leftover GC happens on
// daemon startup instead; each task removes its own work directory on Close.
func Execute() {
	_ = rootCmd.Execute()
}
