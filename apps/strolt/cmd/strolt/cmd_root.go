package cmd

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
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
	Use: "strolt",
	Long: `strolt is a program for backup and restore with
        support for various sources (filesystem, databases), notifications and API.
        Source code is available at https://github.com/strolt/strolt`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		env.Scan()
		metrics.Init()
		scanCliFlags()
	},
}

func Execute() {
	log := logger.New()

	if err := dir.RemoveTempDirectories(); err != nil {
		log.Error(err)
	}

	rootCmd.Execute() //nolint:errcheck

	if err := dir.RemoveTempDirectories(); err != nil {
		log.Error(err)
	}
}
