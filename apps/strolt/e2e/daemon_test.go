package e2e_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	apiUser     = "admin"
	apiPassword = "admin" // pragma: allowlist secret

	snapshotPollTimeout  = 90 * time.Second
	snapshotPollInterval = 2 * time.Second
)

// DaemonSuite exercises the production code path: strolt running as a daemon
// with operations triggered over the HTTP API instead of the CLI.
type DaemonSuite struct {
	suite.Suite

	fs      *Fs
	baseURL string
	client  *http.Client
}

func (s *DaemonSuite) SetupSuite() {
	port, err := containerManager.GetDaemonAPIPort()
	s.Require().NoError(err)

	s.baseURL = "http://localhost:" + port
	s.client = &http.Client{Timeout: 10 * time.Second}

	s.fs = fs()
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(s.fs.createData())
	s.Require().NoError(s.fs.checkValidData())
}

func (s *DaemonSuite) TearDownSuite() {
	s.NoError(s.fs.dropData())
}

func (s *DaemonSuite) TestPing() {
	status, _, err := s.request(http.MethodGet, "/api/v1/ping", false)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, status)
}

func (s *DaemonSuite) TestAPIRequiresAuth() {
	status, _, err := s.request(http.MethodPost, "/api/v1/services/e2e/tasks/daemon/backup", false)
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, status)

	status, _, err = s.request(http.MethodGet, "/api/v1/services/status", false)
	s.Require().NoError(err)
	s.Equal(http.StatusUnauthorized, status)
}

func (s *DaemonSuite) TestBackupViaAPI() {
	status, body, err := s.request(http.MethodPost, "/api/v1/services/e2e/tasks/daemon/backup", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "backup request failed: %s", body)

	// The backup endpoint is asynchronous: poll until the snapshot shows up.
	snapshot, err := s.waitForSnapshot()
	s.Require().NoError(err)

	// Round-trip: wipe the input and restore from the API-created snapshot
	// via the CLI container.
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(strolt("restore", "--service", "e2e", "--task", "daemon", "--destination", "restic-daemon", "--snapshot", snapshot.ID, "--y"))
	s.NoError(s.fs.checkValidData())
}

func (s *DaemonSuite) request(method, path string, withAuth bool) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, nil)
	if err != nil {
		return 0, nil, err
	}

	if withAuth {
		req.SetBasicAuth(apiUser, apiPassword)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)

	return resp.StatusCode, body, err
}

func (s *DaemonSuite) waitForSnapshot() (Snapshot, error) {
	deadline := time.Now().Add(snapshotPollTimeout)

	for {
		snapshots, err := s.snapshotsViaAPI()
		if err == nil && len(snapshots) > 0 {
			return snapshots[0], nil
		}

		if time.Now().After(deadline) {
			return Snapshot{}, fmt.Errorf("no snapshot appeared within %s, last error: %w", snapshotPollTimeout, err)
		}

		time.Sleep(snapshotPollInterval)
	}
}

func (s *DaemonSuite) snapshotsViaAPI() ([]Snapshot, error) {
	status, body, err := s.request(http.MethodGet, "/api/v1/services/e2e/tasks/daemon/destinations/restic-daemon/snapshots", true)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("snapshots request returned %d: %s", status, body)
	}

	var result struct {
		Items []Snapshot `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Items == nil {
		return nil, errors.New("snapshots response has no items")
	}

	return result.Items, nil
}

//nolint:thelper
func DaemonSuiteTest(t *testing.T) {
	tt := timeTook("DaemonSuiteTest")

	suite.Run(t, new(DaemonSuite))
	tt.stop()
}
