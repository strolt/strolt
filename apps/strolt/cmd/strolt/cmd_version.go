package cmd

import (
	gocontext "context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"

	"github.com/spf13/cobra"
)

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

func isExistsVersion(versionList *[]string, version string) bool {
	return slices.Contains(*versionList, version)
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

func finishDriverVersions(versions []string, collectedErrors []error) error {
	binList := excludeDuplicateVersion(&versions)
	if len(binList) == 0 {
		if len(collectedErrors) == 0 {
			Print("binaries not installed")
		}
	} else {
		Print("used binaries:")
		Print(strings.Join(binList, ""))
	}
	return errors.Join(collectedErrors...)
}

func versionContextErrors(processContext gocontext.Context) []error {
	return []error{processContext.Err(), gocontext.Cause(processContext)}
}

func appendVersionContextErrors(processContext gocontext.Context, collectedErrors *[]error) {
	*collectedErrors = append(*collectedErrors, versionContextErrors(processContext)...)
}

func printDriverVersions(processContext gocontext.Context) error {
	var versions []string
	var collectedErrors []error
	c := config.Get()
	finish := func() error { return finishDriverVersions(versions, collectedErrors) }
	stopIfCanceled := func(err error) bool {
		if processContext.Err() != nil {
			appendVersionContextErrors(processContext, &collectedErrors)
			return true
		}
		if errors.Is(err, gocontext.Canceled) || errors.Is(err, gocontext.DeadlineExceeded) {
			collectedErrors = append(collectedErrors, err)
			return true
		}
		return false
	}
	for serviceName, service := range c.Services {
		for taskName, task := range service {
			if processContext.Err() != nil {
				appendVersionContextErrors(processContext, &collectedErrors)
				return finish()
			}
			if collectSourceVersions(processContext, serviceName, taskName, task.Source, &versions, &collectedErrors, stopIfCanceled) {
				return finish()
			}
			for destinationName, destination := range task.Destinations {
				if processContext.Err() != nil {
					appendVersionContextErrors(processContext, &collectedErrors)
					return finish()
				}
				if collectDestinationVersions(processContext, serviceName, taskName, destinationName, destination, &versions, &collectedErrors, stopIfCanceled) {
					return finish()
				}
			}
		}
	}
	if processContext.Err() != nil {
		appendVersionContextErrors(processContext, &collectedErrors)
	}
	return finish()
}

func appendBinaries(versions *[]string, binaries []interfaces.DriverBinaryVersion) {
	for _, bin := range binaries {
		*versions = append(*versions, fmt.Sprintf("%s=%s", bin.Name, bin.Version))
	}
}

func sourceBinaryInventory(processContext gocontext.Context, driver interfaces.DriverSourceInterface) ([]interfaces.DriverBinaryVersion, error) {
	if inventory, ok := driver.(interfaces.DriverSourceBinaryInventoryInterface); ok {
		versions, err := inventory.BinaryVersionInventory(processContext)
		if err != nil {
			return versions, fmt.Errorf("inspect source binary versions: %w", err)
		}
		return versions, nil
	}
	versions, err := driver.BinaryVersion(processContext)
	if err != nil {
		return versions, fmt.Errorf("probe source binary versions: %w", err)
	}
	return versions, nil
}

func collectSourceVersions(processContext gocontext.Context, serviceName, taskName string, source config.DriverSourceConfig, versions *[]string, collectedErrors *[]error, stopIfCanceled func(error) bool) bool {
	if processContext.Err() != nil {
		appendVersionContextErrors(processContext, collectedErrors)
		return true
	}
	driver, err := dmanager.GetSourceDriver(source.Driver, serviceName, taskName, source.Config, source.Env)
	if err != nil {
		*collectedErrors = append(*collectedErrors, fmt.Errorf("service %s task %s source %s: %w", serviceName, taskName, source.Driver, err))
		return stopIfCanceled(err)
	}
	if processContext.Err() != nil {
		appendVersionContextErrors(processContext, collectedErrors)
		return true
	}
	binaries, probeErr := sourceBinaryInventory(processContext, driver)
	appendBinaries(versions, binaries)
	if probeErr == nil {
		return false
	}
	*collectedErrors = append(*collectedErrors, fmt.Errorf("service %s task %s source %s: %w", serviceName, taskName, source.Driver, probeErr))
	return stopIfCanceled(probeErr)
}

func collectDestinationVersions(processContext gocontext.Context, serviceName, taskName, destinationName string, destination config.DriverDestinationConfig, versions *[]string, collectedErrors *[]error, stopIfCanceled func(error) bool) bool {
	if processContext.Err() != nil {
		appendVersionContextErrors(processContext, collectedErrors)
		return true
	}
	driver, err := dmanager.GetDestinationDriver(processContext, destinationName, destination.Driver, serviceName, taskName, destination.Config, destination.Env)
	if err != nil {
		*collectedErrors = append(*collectedErrors, fmt.Errorf("service %s task %s destination %s: %w", serviceName, taskName, destinationName, err))
		return stopIfCanceled(err)
	}
	if processContext.Err() != nil {
		appendVersionContextErrors(processContext, collectedErrors)
		return true
	}
	binaries, probeErr := driver.BinaryVersion(processContext)
	if probeErr == nil {
		appendBinaries(versions, binaries)
		return false
	}
	*collectedErrors = append(*collectedErrors, fmt.Errorf("service %s task %s destination %s: %w", serviceName, taskName, destinationName, probeErr))
	return stopIfCanceled(probeErr)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		Printf("strolt %s compiled with %v on %v/%v\n", ldflags.GetVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
		flag := cmd.Flags()
		if !flag.Changed("config") {
			flag = cmd.InheritedFlags()
		}
		if !flag.Changed("config") {
			return nil
		}
		path, err := flag.GetString("config")
		if err != nil {
			return fmt.Errorf("read config flag: %w", err)
		}
		if path == "" {
			return errors.New("config path must not be empty")
		}
		if err := config.LoadInventory(cmd.Context(), path); err != nil {
			return fmt.Errorf("load version inventory: %w", err)
		}
		return printDriverVersions(cmd.Context())
	},
}
