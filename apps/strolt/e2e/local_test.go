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
	s.NoError(s.fs.dropData())
	s.NoError(s.fs.createData())
	s.NoError(s.fs.checkValidData())
}

func (s *LocalSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

func (s *LocalSuite) TestLocal() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "local", "--y"))

	s.NoError(s.fs.dropData())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "local", "restic-local")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e", "--task", "local", "--destination", "restic-local", "--snapshot", latestSnapshotID, "--y"))
}

func (s *LocalSuite) TestLocalCopy() {
	s.NoError(strolt("backup", "--service", "e2e-copy", "--task", "local", "--y"))

	s.NoError(s.fs.dropData())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-copy", "local", "restic-local")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-copy", "--task", "local", "--destination", "restic-local", "--snapshot", latestSnapshotID, "--y"))
}

func (s *LocalSuite) TestLocalPipe() {
	s.NoError(strolt("backup", "--service", "e2e-pipe", "--task", "local", "--y"))

	s.NoError(s.fs.dropData())

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e-pipe", "local", "restic-local")
	s.NoError(err)

	s.NoError(strolt("restore", "--service", "e2e-pipe", "--task", "local", "--destination", "restic-local", "--snapshot", latestSnapshotID, "--y"))
}

//nolint:thelper
func LocalSuiteTest(t *testing.T) {
	tt := timeTook("LocalSuiteTest")

	suite.Run(t, new(LocalSuite))
	tt.stop()
}
