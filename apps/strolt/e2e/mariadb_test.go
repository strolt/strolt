package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariaDBSuite struct {
	suite.Suite
	c *Conn
}

func (s *MariaDBSuite) SetupSuite() {
	c, err := sqlConnect("mysql", "strolt:strolt@(localhost:9004)/strolt?timeout=60s")
	s.NoError(err)
	s.c = c
}

func (s *MariaDBSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *MariaDBSuite) BeforeTest(suiteName, testName string) {
	s.c.dropTable()
	s.c.createTable()
	s.c.insertData(true)
	s.NoError(s.c.checkValidData())
}

func (s *MariaDBSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MariaDBSuite) TestMariaDB() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "mariadb", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "mariadb", "restic-mariadb")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "mariadb", "--destination", "restic-mariadb", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func MariaDBSuiteTest(t *testing.T) {
	tt := timeTook("MariaDBSuiteTest")

	suite.Run(t, new(MariaDBSuite))
	tt.stop()
}
