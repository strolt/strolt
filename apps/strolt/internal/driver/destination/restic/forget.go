package restic

import (
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

type forgetOutputListItemRemoveListItem struct {
	ID      string    `json:"id"`
	ShortID string    `json:"short_id"`
	Time    time.Time `json:"time"`
	Tags    []string  `json:"tags"`
	Paths   []string  `json:"paths"`
}

type forgetOutputGroupItem struct {
	Remove []forgetOutputListItemRemoveListItem `json:"remove"`
}

type forgetOutput []forgetOutputGroupItem

func (o *forgetOutput) getSnapshotList() []interfaces.Snapshot {
	list := []interfaces.Snapshot{}

	for _, group := range *o {
		for _, remove := range group.Remove {
			list = append(list, interfaces.Snapshot{
				ID:      remove.ID,
				ShortID: remove.ShortID,
				Time:    remove.Time,
				Tags:    remove.Tags,
				Paths:   remove.Paths,
			})
		}
	}

	return list
}
