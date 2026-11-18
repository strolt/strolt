package mysql

import (
	gocontext "context"
	"errors"
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/process"
)

func (i *MySQL) getMySQLVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinMySQL(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mysql version: %w", err)
	}
	arr := strings.Fields(strings.ReplaceAll(string(output), "  ", " "))
	if len(arr) < 3 {
		return "", fmt.Errorf("mysql version: unexpected output %q", string(output))
	}
	return arr[2], nil
}

func (i *MySQL) getMySQLDumpVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinMySQLDump(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mysqldump version: %w", err)
	}
	arr := strings.Fields(strings.ReplaceAll(string(output), "  ", " "))
	if len(arr) < 3 {
		return "", fmt.Errorf("mysqldump version: unexpected output %q", string(output))
	}
	return arr[2], nil
}

// BinaryVersion returns versions of the mysql and mysqldump binaries.
func (i *MySQL) BinaryVersion(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	mysqlVersion, err := i.getMySQLVersion(processContext)
	if err != nil {
		return nil, err
	}
	mysqlDumpVersion, err := i.getMySQLDumpVersion(processContext)
	if err != nil {
		return nil, err
	}
	return []interfaces.DriverBinaryVersion{
		{Name: "mysql", Version: mysqlVersion},
		{Name: "mysqldump", Version: mysqlDumpVersion},
	}, nil
}

// BinaryVersionInventory retains healthy binary versions after probe failures and stops on cancellation.
func (i *MySQL) BinaryVersionInventory(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	var versions []interfaces.DriverBinaryVersion
	var collectedErrors []error
	probes := []struct {
		name  string
		probe func(gocontext.Context) (string, error)
	}{
		{name: "mysql", probe: i.getMySQLVersion},
		{name: "mysqldump", probe: i.getMySQLDumpVersion},
	}
	for _, probe := range probes {
		if err := processContext.Err(); err != nil {
			collectedErrors = append(collectedErrors, err, gocontext.Cause(processContext))
			break
		}
		version, err := probe.probe(processContext)
		if err != nil {
			collectedErrors = append(collectedErrors, err)
			if processContext.Err() != nil || errors.Is(err, gocontext.Canceled) || errors.Is(err, gocontext.DeadlineExceeded) {
				if cause := gocontext.Cause(processContext); cause != nil && !errors.Is(err, cause) {
					collectedErrors = append(collectedErrors, cause)
				}
				break
			}
			continue
		}
		versions = append(versions, interfaces.DriverBinaryVersion{Name: probe.name, Version: version})
	}
	return versions, errors.Join(collectedErrors...)
}
