package e2e_test

import (
	"encoding/json"
	"strings"
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

	_, err := resticExec("s3:http://minio:9000/restic-local", "check")
	s.NoError(err)
}

// TestLocalDestination covers the local destination driver end to end:
// backup into a plain directory, wipe the source, restore, verify content.
func (s *LocalSuite) TestLocalDestination() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "local-dest", "--y"))

	latestSnapshotID, err := stroltGetLatestSnapshotID("e2e", "local-dest", "plain-local")
	s.Require().NoError(err)

	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(strolt("restore", "--service", "e2e", "--task", "local-dest", "--destination", "plain-local", "--snapshot", latestSnapshotID, "--y"))
	s.NoError(s.fs.checkValidData())
}

// TestStatsCLI covers the `strolt stats` command end to end.
func (s *LocalSuite) TestStatsCLI() {
	s.Require().NoError(strolt("backup", "--service", "e2e", "--task", "local", "--y"))

	out, err := stroltWithResponse("stats", "--service", "e2e", "--task", "local", "--destination", "restic-local", "--json")
	s.Require().NoError(err)

	// The CLI echoes "selected ..." lines before the payload.
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	jsonLine := lines[len(lines)-1]

	var stats struct {
		SnapshotsCount int `json:"snapshotsCount"`
	}

	s.Require().NoError(json.Unmarshal([]byte(jsonLine), &stats), "stats output is not valid JSON: %s", out)
	s.Positive(stats.SnapshotsCount)
}

//nolint:thelper
func LocalSuiteTest(t *testing.T) {
	tt := timeTook("LocalSuiteTest")

	suite.Run(t, new(LocalSuite))
	tt.stop()
}
