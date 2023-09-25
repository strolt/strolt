package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresqlSuite struct {
	suite.Suite
	c *Conn
}

func (s *PostgresqlSuite) SetupSuite() {
	c, err := sqlConnect("postgres", "user=strolt password=strolt port=9002 dbname=strolt sslmode=disable connect_timeout=60")
	s.NoError(err)
	s.c = c
}

func (s *PostgresqlSuite) TearDownSuite() {
	s.NoError(s.c.db.Close())
}

func (s *PostgresqlSuite) BeforeTest(suiteName, testName string) {
	s.c.dropTable()
	s.c.createTable()
	s.c.insertData(false)
	s.NoError(s.c.checkValidData())
}

func (s *PostgresqlSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.c.checkValidData())
}

func (s *PostgresqlSuite) TestPostgresql_t() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "pg-t", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "pg-t", "restic-pg-t")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "pg-t", "--destination", "restic-pg-t", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_d() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "pg-d", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "pg-d", "restic-pg-d")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "pg-d", "--destination", "restic-pg-d", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_p() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "pg-p", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "pg-p", "restic-pg-p")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "pg-p", "--destination", "restic-pg-p", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_c() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "pg-c", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "pg-c", "restic-pg-c")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "pg-c", "--destination", "restic-pg-c", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_copy_t() {
	s.NoError(strolt("backup", "--service", "e2e-copy", "--task", "pg-t", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-copy", "pg-t", "restic-pg-t")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-copy", "--task", "pg-t", "--destination", "restic-pg-t", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_copy_d() {
	s.NoError(strolt("backup", "--service", "e2e-copy", "--task", "pg-d", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-copy", "pg-d", "restic-pg-d")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-copy", "--task", "pg-d", "--destination", "restic-pg-d", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_copy_p() {
	s.NoError(strolt("backup", "--service", "e2e-copy", "--task", "pg-p", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-copy", "pg-p", "restic-pg-p")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-copy", "--task", "pg-p", "--destination", "restic-pg-p", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_copy_c() {
	s.NoError(strolt("backup", "--service", "e2e-copy", "--task", "pg-c", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-copy", "pg-c", "restic-pg-c")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-copy", "--task", "pg-c", "--destination", "restic-pg-c", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_pipe_t() {
	s.NoError(strolt("backup", "--service", "e2e-pipe", "--task", "pg-t", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-pipe", "pg-t", "restic-pg-t")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-pipe", "--task", "pg-t", "--destination", "restic-pg-t", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_pipe_d() {
	s.NoError(strolt("backup", "--service", "e2e-pipe", "--task", "pg-d", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-pipe", "pg-d", "restic-pg-d")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-pipe", "--task", "pg-d", "--destination", "restic-pg-d", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_pipe_p() {
	s.NoError(strolt("backup", "--service", "e2e-pipe", "--task", "pg-p", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-pipe", "pg-p", "restic-pg-p")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-pipe", "--task", "pg-p", "--destination", "restic-pg-p", "--snapshot", latestSnapshotID, "--y"))
}

func (s *PostgresqlSuite) TestPostgresql_pipe_c() {
	s.NoError(strolt("backup", "--service", "e2e-pipe", "--task", "pg-c", "--y"))

	s.c.dropTable()

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-pipe", "pg-c", "restic-pg-c")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-pipe", "--task", "pg-c", "--destination", "restic-pg-c", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func PostgresqlSuiteTest(t *testing.T) {
	tt := timeTook("PostgresqlSuiteTest")

	suite.Run(t, new(PostgresqlSuite))
	tt.stop()
}
