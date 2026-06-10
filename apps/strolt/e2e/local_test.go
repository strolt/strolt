package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LocalSuite struct {
	suite.Suite

	fs *Fs
}

func (s *LocalSuite) SetupSuite() {
	s.fs = fs()

	s.NoError(s.fs.dropData())
}

func (s *LocalSuite) TearDownSuite() {
	s.NoError(s.fs.dropData())
}

func (s *LocalSuite) BeforeTest(suiteName, testName string) {
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(s.fs.createData())
	s.Require().NoError(s.fs.checkValidData())
}

func (s *LocalSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

func (s *LocalSuite) TestLocal() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "local", "--y"))

	s.Require().NoError(s.fs.dropData())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "local", "restic-local")
	s.Require().NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "local", "--destination", "restic-local", "--snapshot", latestSnapshotID, "--y"))
}

// TestRepositoryIntegrity runs `restic check` against the repository written
// by the backup: restores can mask repository corruption, an explicit check
// cannot.
func (s *LocalSuite) TestRepositoryIntegrity() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "local", "--y"))

	_, err := execInStrolt("AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin RESTIC_PASSWORD=secret " +
		"/usr/bin/restic -r 's3:http://minio:9000/restic-local' check")
	s.NoError(err)
}

//nolint:thelper
func LocalSuiteTest(t *testing.T) {
	tt := timeTook("LocalSuiteTest")

	suite.Run(t, new(LocalSuite))
	tt.stop()
}
