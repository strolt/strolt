package cmd

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print merged config",
	Long:  `All software has versions. This is Strolt's`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		fmt.Println(config.Yaml()) //nolint:forbidigo
	},
}
