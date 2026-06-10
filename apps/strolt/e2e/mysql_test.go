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
	s.Require().NoError(err)

	// MySQL 8.4 enables TLS with an auto-generated self-signed certificate.
	connStr := fmt.Sprintf("strolt:strolt@(localhost:%s)/strolt?timeout=60s&tls=skip-verify", port)
	c, err := sqlConnect("mysql", connStr)
	s.Require().NoError(err)
	s.c = c
}

func (s *MySQLSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *MySQLSuite) BeforeTest(suiteName, testName string) {
	s.c.dropSchema()
	s.c.createSchema()
	s.c.insertData()
	s.Require().NoError(s.c.checkValidData())
}

func (s *MySQLSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *MySQLSuite) TestMySQL() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "mysql", "restic-mysql")
}

func (s *MySQLSuite) TestMySQL_copy() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "mysql", "restic-mysql")
}

//nolint:thelper
func MySQLSuiteTest(t *testing.T) {
	tt := timeTook("MySQLSuiteTest")

	suite.Run(t, new(MySQLSuite))
	tt.stop()
}
