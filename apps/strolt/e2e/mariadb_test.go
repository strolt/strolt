package e2e_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariaDBSuite struct {
	suite.Suite

	c *Conn
}

func (s *MariaDBSuite) SetupSuite() {
	port, err := containerManager.GetMariaDBPort()
	s.Require().NoError(err)

	connStr := fmt.Sprintf("strolt:strolt@(localhost:%s)/strolt?timeout=60s", port)
	c, err := sqlConnect("mysql", connStr)
	s.Require().NoError(err)
	s.c = c
}

func (s *MariaDBSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *MariaDBSuite) BeforeTest(suiteName, testName string) {
	s.c.dropSchema()
	s.c.createSchema()
	s.c.insertData()
	s.Require().NoError(s.c.checkValidData())
}

func (s *MariaDBSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MariaDBSuite) TestMariaDB() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "mariadb", "restic-mariadb")
}

func (s *MariaDBSuite) TestMariaDB_copy() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "mariadb", "restic-mariadb")
}

//nolint:thelper
func MariaDBSuiteTest(t *testing.T) {
	tt := timeTook("MariaDBSuiteTest")

	suite.Run(t, new(MariaDBSuite))
	tt.stop()
}
