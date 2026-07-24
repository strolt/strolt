package e2e_test

import (
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ForgetSuite struct {
	suite.Suite

	fs *Fs
}

func (s *ForgetSuite) SetupSuite() {
	s.fs = fs()

	s.NoError(s.fs.dropData())
}

func (s *ForgetSuite) TearDownSuite() {
	s.NoError(s.fs.dropData())
}

func (s *ForgetSuite) BeforeTest(suiteName, testName string) {
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(s.fs.createData())
	s.Require().NoError(s.fs.checkValidData())
}

func (s *ForgetSuite) AfterTest(suiteName, testName string) {
	s.NoError(s.fs.checkValidData())
}

// TestForgetSnapshot deletes a snapshot the retention policy (keep last=3)
// would preserve, which is what separates forget from prune.
func (s *ForgetSuite) TestForgetSnapshot() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "forget", "restic-forget")
	s.Require().NoError(err)
	s.Require().Len(snapshots, 2)

	removed := snapshots[0].ID

	s.Require().NoError(strolt("forget",
		"--service", "e2e", "--task", "forget", "--destination", "restic-forget", "--snapshot", removed, "--y"))

	snapshots, err = stroltGetSnapshotList("e2e", "forget", "restic-forget")
	s.Require().NoError(err)
	s.Require().Len(snapshots, 1)
	s.NotEqual(removed, snapshots[0].ID)

	// The data of the forgotten snapshot is pruned, so the repository must stay
	// consistent afterwards.
	_, err = resticExec("s3:http://minio:9000/restic-forget", "check")
	s.NoError(err)
}

// TestForgetSnapshotNoLock covers a destination configured with 'no-lock: true':
// restic aborts a real forget when --no-lock is passed, so strolt has to drop
// the flag for it.
func (s *ForgetSuite) TestForgetSnapshotNoLock() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "forget", "restic-forget-nolock")
	s.Require().NoError(err)
	s.Require().NotEmpty(snapshots)

	removed := snapshots[0].ID

	s.Require().NoError(strolt("forget",
		"--service", "e2e", "--task", "forget", "--destination", "restic-forget-nolock", "--snapshot", removed, "--y"))

	snapshots, err = stroltGetSnapshotList("e2e", "forget", "restic-forget-nolock")
	s.Require().NoError(err)

	for _, snapshot := range snapshots {
		s.NotEqual(removed, snapshot.ID)
	}
}

// TestForgetUnknownSnapshot guards the case restic itself is silent about:
// forgetting an unknown ID exits with status 0 and only warns on stderr.
func (s *ForgetSuite) TestForgetUnknownSnapshot() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	err := strolt("forget", "--service", "e2e", "--task", "forget",
		"--destination", "restic-forget", "--snapshot", "0000000000000000000000000000000000000000000000000000000000000000", "--y")
	s.Error(err)
}

// operationTypePattern extracts the operation type from the context logged by
// the console notification driver. Text-formatted log lines escape the JSON
// payload, so the quotes around the value are optional.
var operationTypePattern = regexp.MustCompile(`opertationType\\?":\\?"([A-Z]+)`)

func notifiedOperationTypes(output []byte) []string {
	types := []string{}

	for _, match := range operationTypePattern.FindAllStringSubmatch(string(output), -1) {
		if !slices.Contains(types, match[1]) {
			types = append(types, match[1])
		}
	}

	return types
}

// TestNotificationOperationType guards the operation type a command reports to
// its notification drivers: prune used to announce itself as RESTORE.
func (s *ForgetSuite) TestNotificationOperationType() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	output, err := stroltWithResponse("prune",
		"--service", "e2e", "--task", "forget", "--destination", "restic-forget", "--y")
	s.Require().NoError(err)
	s.Equal([]string{"PRUNE"}, notifiedOperationTypes(output))

	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	snapshots, err := stroltGetSnapshotList("e2e", "forget", "restic-forget")
	s.Require().NoError(err)
	s.Require().NotEmpty(snapshots)

	output, err = stroltWithResponse("forget",
		"--service", "e2e", "--task", "forget", "--destination", "restic-forget", "--snapshot", snapshots[0].ID, "--y")
	s.Require().NoError(err)
	s.Equal([]string{"FORGET"}, notifiedOperationTypes(output))
}

func (s *ForgetSuite) TestUnlock() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))

	s.Require().NoError(strolt("unlock", "--service", "e2e", "--task", "forget", "--destination", "restic-forget"))
	s.Require().NoError(strolt("unlock", "--service", "e2e", "--task", "forget", "--destination", "restic-forget", "--remove-all"))

	// Unlocking must leave the repository usable.
	s.NoError(strolt("backup", "--service", "e2e", "--task", "forget", "--y"))
}

//nolint:thelper
func ForgetSuiteTest(t *testing.T) {
	tt := timeTook("ForgetSuiteTest")

	suite.Run(t, new(ForgetSuite))
	tt.stop()
}
