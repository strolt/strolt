package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print result config",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		logger.New().Infof(config.Yaml())
	},
}
