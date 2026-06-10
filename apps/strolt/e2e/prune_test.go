package e2e_test

import (
	"fmt"
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
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(s.fs.createData())
	s.Require().NoError(s.fs.checkValidData())
}

func (s *PruneSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

// TestRetentionTimeBuckets verifies hourly/daily keep policies. The snapshot
// timestamps are pinned in the past via `restic backup --time`, so the bucket
// assignment never races with the wall clock and the expectation is stable.
//
// Policy: keep hourly=2, daily=1. Hours with snapshots, newest first:
// 2020-01-02 09h (keeps 09:15), 2020-01-01 11h (keeps 11:30); daily keeps
// 2020-01-02 09:15 again. Everything else must be pruned.
func (s *PruneSuite) TestRetentionTimeBuckets() {
	repository := "s3:http://minio:9000/restic-retention"

	backdated := []string{
		"2020-01-01 10:00:00",
		"2020-01-01 11:00:00",
		"2020-01-01 11:30:00",
		"2020-01-02 09:00:00",
		"2020-01-02 09:15:00",
	}

	for _, timestamp := range backdated {
		_, err := resticExec(repository, fmt.Sprintf("backup --time '%s' /e2e/input", timestamp))
		s.Require().NoError(err)
	}

	snapshots, err := stroltGetSnapshotList("e2e", "retention", "restic-retention")
	s.Require().NoError(err)
	s.Require().Len(snapshots, len(backdated))

	s.Require().NoError(strolt("prune", "--service", "e2e", "--task", "retention", "--destination", "restic-retention", "--y"))

	snapshots, err = stroltGetSnapshotList("e2e", "retention", "restic-retention")
	s.Require().NoError(err)

	kept := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		kept = append(kept, snapshot.Date)
	}

	s.ElementsMatch([]string{"2020-01-02T09:15:00Z", "2020-01-01T11:30:00Z"}, kept)
}

func (s *PruneSuite) TestPruneRestic() {
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))
	s.NoError(strolt("backup", "--service", "e2e", "--task", "prune", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "prune", "restic-prune")
	s.Require().NoError(err)
	s.Len(snapshots, 5)

	log.Println("before:", snapshots)

	s.Require().NoError(strolt("prune", "--service", "e2e", "--task", "prune", "--destination", "restic-prune", "--y"))

	snapshots, err = stroltGetSnapshotList("e2e", "prune", "restic-prune")
	s.Require().NoError(err)

	log.Println("after:", snapshots)

	s.Len(snapshots, 3)
}

//nolint:thelper
func PruneSuiteTest(t *testing.T) {
	tt := timeTook("PruneSuiteTest")

	suite.Run(t, new(PruneSuite))
	tt.stop()
}
