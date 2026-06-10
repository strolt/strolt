package e2e_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const negativeConfigPath = "/strolt/.strolt/negative-config.yml"

// stroltWithConfig runs the strolt CLI against an alternative config file.
func stroltWithConfig(configPath string, args ...string) error {
	_, err := execInStrolt("/strolt/bin/strolt --config " + configPath + " " + strings.Join(args, " "))

	return err
}

// NegativeSuite verifies that strolt fails loudly (non-zero exit code) when
// the environment is broken, instead of producing empty or bogus backups.
type NegativeSuite struct {
	suite.Suite
}

func (s *NegativeSuite) TestWrongResticPassword() {
	s.Require().Error(stroltWithConfig(negativeConfigPath,
		"snapshots", "--service", "negative", "--task", "bad-password", "--destination", "restic-bad-password", "--json"))

	s.Require().Error(stroltWithConfig(negativeConfigPath,
		"backup", "--service", "negative", "--task", "bad-password", "--y"))
}

func (s *NegativeSuite) TestBackupUnreachableDatabase() {
	s.Require().Error(stroltWithConfig(negativeConfigPath,
		"backup", "--service", "negative", "--task", "bad-host", "--y"))
}

func (s *NegativeSuite) TestPipeModeUnsupportedDriver() {
	// The mysql driver does not implement streaming backups; pipe mode must
	// be rejected instead of silently falling back to copy.
	s.Require().Error(stroltWithConfig(negativeConfigPath,
		"backup", "--service", "negative", "--task", "pipe-unsupported", "--y"))
}

func (s *NegativeSuite) TestRestoreNonexistentSnapshot() {
	s.Require().Error(strolt("restore", "--service", "e2e", "--task", "local", "--destination", "restic-local", "--snapshot", "deadbeef", "--y"))
}

//nolint:thelper
func NegativeSuiteTest(t *testing.T) {
	tt := timeTook("NegativeSuiteTest")

	suite.Run(t, new(NegativeSuite))
	tt.stop()
}
