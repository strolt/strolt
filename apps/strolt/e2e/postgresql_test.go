package e2e_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresqlSuite struct {
	suite.Suite

	c *Conn
}

func (s *PostgresqlSuite) SetupSuite() {
	port, err := containerManager.GetPostgresPort()
	s.Require().NoError(err)

	connStr := fmt.Sprintf("user=strolt password=strolt host=localhost port=%s dbname=strolt sslmode=disable connect_timeout=60", port)
	c, err := sqlConnect("postgres", connStr)
	s.Require().NoError(err)
	s.c = c
}

func (s *PostgresqlSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *PostgresqlSuite) BeforeTest(suiteName, testName string) {
	s.c.dropSchema()
	s.c.createSchema()
	s.c.insertData()
	s.Require().NoError(s.c.checkValidData())
}

func (s *PostgresqlSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *PostgresqlSuite) TestPostgresql_t() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "pg-t", "restic-pg-t")
}

func (s *PostgresqlSuite) TestPostgresql_d() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "pg-d", "restic-pg-d")
}

func (s *PostgresqlSuite) TestPostgresql_p() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "pg-p", "restic-pg-p")
}

func (s *PostgresqlSuite) TestPostgresql_c() {
	sqlRoundTrip(&s.Suite, s.c, "e2e", "pg-c", "restic-pg-c")
}

func (s *PostgresqlSuite) TestPostgresql_copy_t() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "pg-t", "restic-pg-t")
}

func (s *PostgresqlSuite) TestPostgresql_copy_d() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "pg-d", "restic-pg-d")
}

func (s *PostgresqlSuite) TestPostgresql_copy_p() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "pg-p", "restic-pg-p")
}

func (s *PostgresqlSuite) TestPostgresql_copy_c() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-copy", "pg-c", "restic-pg-c")
}

func (s *PostgresqlSuite) TestPostgresql_pipe_t() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-pipe", "pg-t", "restic-pg-t")
}

func (s *PostgresqlSuite) TestPostgresql_pipe_d() {
	s.T().Skip("PostgreSQL directory format (d) does not support pipe mode - this is expected behavior")
}

func (s *PostgresqlSuite) TestPostgresql_pipe_p() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-pipe", "pg-p", "restic-pg-p")
}

func (s *PostgresqlSuite) TestPostgresql_pipe_c() {
	sqlRoundTrip(&s.Suite, s.c, "e2e-pipe", "pg-c", "restic-pg-c")
}

//nolint:thelper
func PostgresqlSuiteTest(t *testing.T) {
	tt := timeTook("PostgresqlSuiteTest")

	suite.Run(t, new(PostgresqlSuite))
	tt.stop()
}
