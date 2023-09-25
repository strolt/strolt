package task

import "github.com/strolt/strolt/apps/strolt/internal/dmanager"

func (t *Task) IsSourceEmpty() (bool, error) {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return false, err
	}

	return sourceDriver.IsEmpty()
}
