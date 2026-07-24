package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const restoreTargetPath = "/e2e/restore-target"

type TuningSuite struct {
	suite.Suite

	fs *Fs
}

func (s *TuningSuite) SetupSuite() {
	s.fs = fs()

	s.NoError(s.fs.dropData())
}

func (s *TuningSuite) TearDownSuite() {
	s.NoError(s.fs.dropData())
}

func (s *TuningSuite) BeforeTest(suiteName, testName string) {
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(s.fs.createData())
	s.Require().NoError(s.fs.checkValidData())

	_, err := execInStrolt("rm -rf " + restoreTargetPath)
	s.Require().NoError(err)
}

func (s *TuningSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

// TestBackupAndRestoreWithTuning runs a full cycle against a destination that
// sets every tuning option at once, including a RESTIC_PASSWORD passed only
// through env_extra: restic rejects the repository if those pairs do not reach
// the process.
func (s *TuningSuite) TestBackupAndRestoreWithTuning() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "tuning", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "tuning", "restic-tuning")
	s.Require().NoError(err)
	s.Require().NotEmpty(snapshots)

	s.Require().NoError(strolt("restore",
		"--service", "e2e", "--task", "tuning", "--destination", "restic-tuning", "--snapshot", snapshots[0].ID, "--y"))

	// restore-target keeps restic out of the source: the snapshot content is
	// written into the configured directory instead, with the same layout it
	// would have had in the source.
	output, err := execInStrolt("find " + restoreTargetPath + " -type f | sort")
	s.Require().NoError(err)

	for _, name := range []string{"0.txt", "1/1.txt", "bin/seq.txt"} {
		s.Contains(string(output), restoreTargetPath+"/"+name)
	}

	_, err = resticExec("s3:http://minio:9000/restic-tuning", "check")
	s.NoError(err)
}

//nolint:thelper
func TuningSuiteTest(t *testing.T) {
	tt := timeTook("TuningSuiteTest")

	suite.Run(t, new(TuningSuite))
	tt.stop()
}
