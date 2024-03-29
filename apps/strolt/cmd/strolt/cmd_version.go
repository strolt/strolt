package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

func isExistsVersion(versionList *[]string, version string) bool {
	for _, _version := range *versionList {
		if _version == version {
			return true
		}
	}

	return false
}

func excludeDuplicateVersion(versionList *[]string) []string {
	var newVersionList []string

	for _, version := range *versionList {
		if !isExistsVersion(&newVersionList, version) {
			newVersionList = append(newVersionList, version)
		}
	}

	return newVersionList
}

func printDriverVersions() {
	var versions []string

	c := config.Get()

	for serviceName, service := range c.Services {
		for taskName, task := range service {
			d, err := dmanager.GetSourceDriver(task.Source.Driver, serviceName, taskName, task.Source.Config, task.Source.Env)
			if err == nil {
				binList, _ := d.BinaryVersion()

				for _, bin := range binList {
					versions = append(versions, fmt.Sprintf("%s=%s", bin.Name, bin.Version))
				}
			}

			for destinationName, destination := range task.Destinations {
				d, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, serviceName, taskName, destination.Config, destination.Env)
				if err == nil {
					binList, _ := d.BinaryVersion()

					for _, bin := range binList {
						versions = append(versions, fmt.Sprintf("%s=%s", bin.Name, bin.Version))
					}
				}
			}
		}
	}

	binList := excludeDuplicateVersion(&versions)

	if len(binList) == 0 {
		Print("binaries not installed")
	} else {
		Print("used binaries:")
		Print(strings.Join(binList, ""))
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		Printf("strolt %s compiled with %v on %v/%v\n", ldflags.GetVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)

		printDriverVersions()
	},
}
