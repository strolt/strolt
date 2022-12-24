package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/config"
)

var (
	cliConfig config.CliConfig
)

//nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().StringSliceVar(&cliConfig.Tags, "tag", []string{}, "set global tag. Example: (--tag tag_one --tag tag_two)")
}

func scanCliFlags() {
	config.SetCliConfig(&cliConfig)
}
