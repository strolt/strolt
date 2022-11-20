package e2e_test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PruneSuite struct {
	suite.Suite
	fs *Fs
}

func (s *PruneSuite) SetupSuite() {
	s.fs = fs()

	s.NoError(s.fs.dropData())
}

func (s *PruneSuite) TearDownSuite() {
	s.NoError(s.fs.dropData())
}

func (s *PruneSuite) BeforeTest(suiteName, testName string) {
	s.NoError(s.fs.dropData())
	s.NoError(s.fs.createData())
	s.NoError(s.fs.checkValidData())
}

func (s *PruneSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

func (s *PruneSuite) TestPruneRestic() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "prune", "restic-prune")
	s.NoError(err)
	s.Equal(len(snapshots), 5)

	log.Println("before:", snapshots)

	s.NoError(strolt("prune", "--service", "e2e", "--task", "prune", "--destination", "restic-prune", "--y"))

	snapshots, err = stroltGetSnapshotList("e2e", "prune", "restic-prune")
	s.NoError(err)

	log.Println("after:", snapshots)

	s.Equal(len(snapshots), 3)
}

//nolint:thelper
func PruneSuiteTest(t *testing.T) {
	tt := timeTook("PruneSuiteTest")

	suite.Run(t, new(PruneSuite))
	tt.stop()
}
