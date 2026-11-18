package mongodb

import (
	gocontext "context"
	"errors"
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/process"
)

func (i *MongoDB) getMongoDumpVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinMongoDump(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mongodump version: %w", err)
	}
	str := string(output)
	arr := strings.Fields(strings.SplitN(str, "\n", 2)[0])
	if len(arr) < 3 || arr[2] == "" {
		return "", fmt.Errorf("mongodump version: unexpected output %q", str)
	}
	return arr[2], nil
}

func (i *MongoDB) getMongoRestoreVersion(processContext gocontext.Context) (string, error) {
	cmd := process.CommandContext(processContext, i.getBinMongoRestore(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mongorestore version: %w", err)
	}
	str := string(output)
	arr := strings.Fields(strings.SplitN(str, "\n", 2)[0])
	if len(arr) < 3 || arr[2] == "" {
		return "", fmt.Errorf("mongorestore version: unexpected output %q", str)
	}
	return arr[2], nil
}

// BinaryVersion returns versions of the mongodump and mongorestore binaries.
func (i *MongoDB) BinaryVersion(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	mongodumpVersion, err := i.getMongoDumpVersion(processContext)
	if err != nil {
		return nil, err
	}
	mongorestoreVersion, err := i.getMongoRestoreVersion(processContext)
	if err != nil {
		return nil, err
	}
	return []interfaces.DriverBinaryVersion{
		{Name: "mongodump", Version: mongodumpVersion},
		{Name: "mongorestore", Version: mongorestoreVersion},
	}, nil
}

// BinaryVersionInventory retains healthy binary versions after probe failures and stops on cancellation.
func (i *MongoDB) BinaryVersionInventory(processContext gocontext.Context) ([]interfaces.DriverBinaryVersion, error) {
	var versions []interfaces.DriverBinaryVersion
	var collectedErrors []error
	probes := []struct {
		name  string
		probe func(gocontext.Context) (string, error)
	}{
		{name: "mongodump", probe: i.getMongoDumpVersion},
		{name: "mongorestore", probe: i.getMongoRestoreVersion},
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
