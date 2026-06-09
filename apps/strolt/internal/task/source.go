package task

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
)

// IsSourceEmpty reports whether the task source contains no data.
func (t *Task) IsSourceEmpty() (bool, error) {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return false, fmt.Errorf("get source driver: %w", err)
	}

	isEmpty, err := sourceDriver.IsEmpty()
	if err != nil {
		return false, fmt.Errorf("check source is empty: %w", err)
	}

	return isEmpty, nil
}
