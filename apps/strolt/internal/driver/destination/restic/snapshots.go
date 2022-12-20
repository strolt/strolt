package restic

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

type resticSnapshot struct {
	Time     time.Time `json:"time"`
	Tree     string    `json:"tree"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	UID      int       `json:"uid"`
	GID      int       `json:"gid"`
	ID       string    `json:"id"`
	ShortID  string    `json:"short_id"`
	Tags     []string  `json:"tags"`
}

func (i *Restic) Snapshots() ([]interfaces.Snapshot, error) {
	cmd := exec.Command(i.getBin(), "--json", "snapshots")

	env, err := i.getEnv()
	if err != nil {
		return nil, err
	}

	cmd.Env = env

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)

	i.logger.Debug(string(output))

	if err != nil {
		i.logger.Error(err)
		return nil, err
	}

	var resticSnapshots []resticSnapshot
	if err := json.Unmarshal(output, &resticSnapshots); err != nil {
		i.logger.Error(err)
		return nil, err
	}

	snapshots := []interfaces.Snapshot{}

	for _, snapshot := range resticSnapshots {
		snapshots = append(snapshots, interfaces.Snapshot{
			ID:      snapshot.ID,
			ShortID: snapshot.ShortID,
			Time:    snapshot.Time,
			Tags:    snapshot.Tags,
			Paths:   snapshot.Paths,
		})
	}

	return snapshots, nil
}
