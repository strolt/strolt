package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

// TestPostgresql_pipe_restore_conflict pins the pipe-deadlock regression: a
// piped restore whose consumer aborts early (here: restoring over an existing
// schema makes pg_restore/psql fail on the first conflicting statement) must
// return an error promptly. It used to hang forever with zero CPU because the
// task waited for the streaming side while nobody drained its pipe anymore.
func (s *PostgresqlSuite) TestPostgresql_pipe_restore_conflict() {
	serviceName, taskName, destinationName := "e2e-pipe", "pg-c", "restic-pg-c"

	// The fixture dump is tiny and would fit into the kernel pipe buffers:
	// the streaming side would then exit on its own before the consumer
	// aborts, and the deadlock would never trigger even on broken code. Pad
	// the database so the dump is orders of magnitude larger than any pipe
	// buffer while the restore consumer still dies on the first conflict.
	// dropSchema() does not know this table, so the test removes it itself.
	s.c.db.MustExec(`CREATE TABLE e2e_pipe_conflict_blob (payload text NOT NULL)`)
	defer s.c.dropBlobTable()

	s.c.db.MustExec(`INSERT INTO e2e_pipe_conflict_blob SELECT string_agg(md5(random()::text), '') FROM generate_series(1, 2000000)`)

	s.Require().NoError(strolt("backup", "--service", serviceName, "--task", taskName, "--y"))

	latestSnapshotID, err := stroltGetLatestSnapshotID(serviceName, taskName, destinationName)
	s.Require().NoError(err)

	// The schema still exists (BeforeTest created it), so the restore
	// consumer dies on the first "already exists" error mid-stream.
	bounded, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	out, err := execInStroltCtx(bounded, fmt.Sprintf(
		"/strolt/bin/strolt restore --service %s --task %s --destination %s --snapshot %s --y",
		serviceName, taskName, destinationName, latestSnapshotID,
	))
	s.Require().Error(err, "restoring over an existing schema must fail, not report success")
	s.Require().NoError(bounded.Err(), "the restore must fail on its own; hitting the test deadline means the pipe deadlock is back")
	s.Require().Contains(err.Error(), "exited with code", "the CLI must have run to a non-zero exit instead of being cut off mid-run")
	s.Require().Contains(string(out), "already exists",
		"the piped restore must surface the pg_restore/psql SQL error that aborted it, not just the exit status")
}

// dropBlobTable removes the padding table created by
// TestPostgresql_pipe_restore_conflict. dropSchema does not know it, and a
// leftover table would make every later restore in the suite fail on the
// same "already exists" conflict.
func (c *Conn) dropBlobTable() {
	c.db.MustExec(`DROP TABLE IF EXISTS e2e_pipe_conflict_blob;`)
}

//nolint:thelper
func PostgresqlSuiteTest(t *testing.T) {
	tt := timeTook("PostgresqlSuiteTest")

	suite.Run(t, new(PostgresqlSuite))
	tt.stop()
}
