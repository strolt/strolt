package mongodb

import (
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *MongoDB) getMongoDumpVersion() string {
	cmd := exec.Command(i.getBinMongoDump(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	str := string(output)
	arr := strings.Split(str, "\n")

	if len(arr) != 0 {
		arr := strings.Split(arr[0], " ")
		return arr[2]
	}

	return str
}

func (i *MongoDB) getMongoRestoreVersion() string {
	cmd := exec.Command(i.getBinMongoRestore(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	str := string(output)
	arr := strings.Split(str, "\n")

	if len(arr) != 0 {
		arr := strings.Split(arr[0], " ")
		return arr[2]
	}

	return str
}

func (i *MongoDB) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	mongodumpVersion := i.getMongoDumpVersion()
	mongorestoreVersion := i.getMongoRestoreVersion()

	return []interfaces.DriverBinaryVersion{
		{Name: "mongodump", Version: mongodumpVersion},
		{Name: "mongorestore", Version: mongorestoreVersion},
	}, nil
}
