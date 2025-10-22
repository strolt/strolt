package e2e_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MySQLSuite struct {
	suite.Suite
	c *Conn
}

func (s *MySQLSuite) SetupSuite() {
	port, err := containerManager.GetMySQLPort()
	s.NoError(err)

	connStr := fmt.Sprintf("strolt:strolt@(localhost:%s)/strolt?timeout=60s", port)
	c, err := sqlConnect("mysql", connStr)
	s.NoError(err)
	s.c = c
}

func (s *MySQLSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *MySQLSuite) BeforeTest(suiteName, testName string) {
	s.c.dropTable()
	s.c.createTable()
	s.c.insertData(true)
	s.NoError(s.c.checkValidData())
}

func (s *MySQLSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MySQLSuite) TestMySQL() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "mysql", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "mysql", "restic-mysql")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "mysql", "--destination", "restic-mysql", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func MySQLSuiteTest(t *testing.T) {
	tt := timeTook("MySQLSuiteTest")

	suite.Run(t, new(MySQLSuite))
	tt.stop()
}
