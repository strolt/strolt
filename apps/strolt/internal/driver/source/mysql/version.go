package mysql

import (
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/process"
)

func (i *MySQL) getMySQLVersion() string {
	cmd := process.Command(i.getBinMySQL(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}
	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")
	return arr[2]
}

func (i *MySQL) getMySQLDumpVersion() string {
	cmd := process.Command(i.getBinMySQLDump(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}
	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")
	return arr[2]
}

// BinaryVersion returns versions of the mysql and mysqldump binaries.
func (i *MySQL) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	mysqlVersion := i.getMySQLVersion()
	mysqlDumpVersion := i.getMySQLDumpVersion()

	return []interfaces.DriverBinaryVersion{
		{Name: "mysql", Version: mysqlVersion},
		{Name: "mysqldump", Version: mysqlDumpVersion},
	}, nil
}
