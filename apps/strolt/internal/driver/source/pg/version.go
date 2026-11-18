package pg

import (
	gocontext "context"
	"errors"
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/process"
)

func (i *PgDump) getPgDumpVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinPgDump(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pg_dump version: %w", err)
	}
	arr := strings.Split(string(output), " ")
	if len(strings.Fields(string(output))) < 3 || strings.TrimSpace(arr[len(arr)-1]) == "" {
		return "", fmt.Errorf("pg_dump version: unexpected output %q", string(output))
	}
	return arr[len(arr)-1], nil
}

func (i *PgDump) getPgRestoreVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinPgRestore(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pg_restore version: %w", err)
	}
	arr := strings.Split(string(output), " ")
	if len(strings.Fields(string(output))) < 3 || strings.TrimSpace(arr[len(arr)-1]) == "" {
		return "", fmt.Errorf("pg_restore version: unexpected output %q", string(output))
	}
	return arr[len(arr)-1], nil
}

func (i *PgDump) getPsqlVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinPSQL(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("psql version: %w", err)
	}
	arr := strings.Split(string(output), " ")
	if len(strings.Fields(string(output))) < 3 || strings.TrimSpace(arr[len(arr)-1]) == "" {
		return "", fmt.Errorf("psql version: unexpected output %q", string(output))
	}
	return arr[len(arr)-1], nil
}

// BinaryVersion returns versions of the pg_dump, pg_restore and psql binaries.
func (i *PgDump) BinaryVersion(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	pgDumpVersion, err := i.getPgDumpVersion(processContext)
	if err != nil {
		return nil, err
	}
	pgRestoreVersion, err := i.getPgRestoreVersion(processContext)
	if err != nil {
		return nil, err
	}
	pgPsqlVersion, err := i.getPsqlVersion(processContext)
	if err != nil {
		return nil, err
	}
	return []interfaces.DriverBinaryVersion{
		{Name: "pg_dump", Version: pgDumpVersion},
		{Name: "pg_restore", Version: pgRestoreVersion},
		{Name: "psql", Version: pgPsqlVersion},
	}, nil
}

// BinaryVersionInventory retains healthy binary versions after probe failures and stops on cancellation.
func (i *PgDump) BinaryVersionInventory(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	var versions []interfaces.DriverBinaryVersion
	var collectedErrors []error
	probes := []struct {
		name  string
		probe func(gocontext.Context) (string, error)
	}{
		{name: "pg_dump", probe: i.getPgDumpVersion},
		{name: "pg_restore", probe: i.getPgRestoreVersion},
		{name: "psql", probe: i.getPsqlVersion},
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
