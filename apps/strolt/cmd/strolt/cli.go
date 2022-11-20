package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/config"
)

var (
	cliConfig config.CliConfig
)

//nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.Tags, "config.tags", []string{}, "global destination tags")
}

func parseCliFlags() {
	config.SetCliConfig(&cliConfig)
}
