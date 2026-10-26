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
	snapshot := s.backupViaAPIAndAwait()

	// Round-trip: wipe the input and restore from the API-created snapshot
	// via the CLI container.
	s.Require().NoError(s.fs.dropData())
	s.Require().NoError(strolt("restore", "--service", "e2e", "--task", "daemon", "--destination", "restic-daemon", "--snapshot", snapshot.ID, "--y"))
	s.NoError(s.fs.checkValidData())
}

func (s *DaemonSuite) TestInfoViaAPI() {
	status, body, err := s.request(http.MethodGet, "/api/v1/info", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "info request failed: %s", body)

	var info struct {
		Version         string `json:"version"`
		StartedAt       string `json:"startedAt"`
		ConfigUpdatedAt string `json:"configUpdatedAt"`
	}

	s.Require().NoError(json.Unmarshal(body, &info), "info payload is not valid JSON: %s", body)
	s.NotEmpty(info.Version)
	s.NotEmpty(info.StartedAt)
}

func (s *DaemonSuite) TestConfigViaAPI() {
	status, body, err := s.request(http.MethodGet, "/api/v1/config", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "config request failed: %s", body)

	var cfg map[string]any

	s.Require().NoError(json.Unmarshal(body, &cfg), "config payload is not valid JSON: %s", body)
	s.Contains(cfg, "services")
}

func (s *DaemonSuite) TestStatusViaAPI() {
	status, body, err := s.request(http.MethodGet, "/api/v1/services/status", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "status request failed: %s", body)
	s.Contains(string(body), "e2e")
}

func (s *DaemonSuite) TestStatsViaAPI() {
	// Self-contained: create this test's own snapshot instead of relying on
	// other tests (testify does not guarantee declaration order).
	s.backupViaAPIAndAwait()

	status, body, err := s.request(http.MethodGet, "/api/v1/services/e2e/tasks/daemon/destinations/restic-daemon/stats", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "stats request failed: %s", body)

	var result struct {
		Data struct {
			SnapshotsCount int `json:"snapshotsCount"`
		} `json:"data"`
	}

	s.Require().NoError(json.Unmarshal(body, &result), "stats payload is not valid JSON: %s", body)
	s.Positive(result.Data.SnapshotsCount)
}

// TestPruneViaAPI seeds more snapshots than the fixture's keep.last: 3
// retention allows, checks the preview lists exactly the surplus seeded
// snapshots and verifies that the executed prune actually reduced the
// repository to the retained set.
func (s *DaemonSuite) TestPruneViaAPI() {
	const (
		seededSnapshots = 5
		retainedCount   = 3
	)

	seeded := map[string]bool{}

	for range seededSnapshots {
		snapshot := s.backupViaAPIAndAwait()
		seeded[snapshot.ID] = true
	}

	// The repository is shared across the whole suite, so the prune surplus
	// contains this test's seeds plus any earlier tests' snapshots.
	before := s.snapshotIDs()

	totalBefore := len(before)

	status, body, err := s.request(http.MethodGet, "/api/v1/services/e2e/tasks/daemon/destinations/restic-daemon/snapshots/prune", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "snapshots-for-prune request failed: %s", body)

	var preview struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	s.Require().NoError(json.Unmarshal(body, &preview), "prune preview payload is not valid JSON: %s", body)
	s.Require().Len(preview.Data, totalBefore-retainedCount, "prune preview candidates: %v", preview.Data)

	for _, candidate := range preview.Data {
		s.True(before[candidate.ID], "preview candidate %s did not exist before the prune", candidate.ID)
	}

	status, body, err = s.request(http.MethodPost, "/api/v1/services/e2e/tasks/daemon/destinations/restic-daemon/prune", true)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, status, "prune request failed: %s", body)

	s.Eventually(func() bool {
		return len(s.snapshotIDs()) == retainedCount
	}, snapshotPollTimeout, snapshotPollInterval, "expected %d snapshots after prune", retainedCount)
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

// backupViaAPIAndAwait triggers a backup over the API and waits until a
// snapshot that did not exist before the POST shows up. Retrying POSTs and
// waiting for a new ID are both required: the API rejects requests with 400
// while the task is still running, and pre-existing snapshots would
// otherwise make an ID check pass instantly.
func (s *DaemonSuite) backupViaAPIAndAwait() Snapshot {
	before := s.snapshotIDs()

	var status int
	var body []byte

	s.Eventually(func() bool {
		var err error

		status, body, err = s.request(http.MethodPost, "/api/v1/services/e2e/tasks/daemon/backup", true)
		if err != nil {
			return false
		}

		if status == http.StatusBadRequest {
			return false // task still busy with a previous operation
		}

		return true
	}, snapshotPollTimeout, snapshotPollInterval, "backup POST never succeeded, last status %d: %s", status, body)

	deadline := time.Now().Add(snapshotPollTimeout)

	for {
		// Snapshot listing is itself a task and returns 400 while the backup
		// is running, so treat listing errors as "not ready yet".
		snapshots, err := s.snapshotsViaAPI()
		if err == nil {
			for _, snapshot := range snapshots {
				if !before[snapshot.ID] {
					return snapshot
				}
			}
		}

		if time.Now().After(deadline) {
			s.Require().Failf("no new snapshot appeared within %s", snapshotPollTimeout.String())
		}

		time.Sleep(snapshotPollInterval)
	}
}

func (s *DaemonSuite) snapshotIDs() map[string]bool {
	snapshots, err := s.snapshotsViaAPI()
	s.Require().NoError(err)

	ids := make(map[string]bool, len(snapshots))
	for _, snapshot := range snapshots {
		ids[snapshot.ID] = true
	}

	return ids
}

//nolint:thelper
func DaemonSuiteTest(t *testing.T) {
	tt := timeTook("DaemonSuiteTest")

	suite.Run(t, new(DaemonSuite))
	tt.stop()
}
