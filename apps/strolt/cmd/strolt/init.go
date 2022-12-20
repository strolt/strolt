package cmd

import (
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/shared/logger"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize destinations",
	Long:  `All software has versions. This is Strolt's`,
	Run: func(cmd *cobra.Command, args []string) {

		initConfig()

		c := config.Get()

		for serviceName, service := range c.Services {
			for taskName, task := range service {
				for destinationName, destination := range task.Destinations {
					log := logger.New().WithField("destination", destinationName)
					log.Info("try init")
					d, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, serviceName, taskName, destination.Config, destination.Env)
					if err != nil {
						log.Error(err)
					} else {
						if err := d.Init(); err != nil {
							log.Error(err)
						} else {
							log.Info("initialized")
						}
					}
				}
			}
		}
	},
}
