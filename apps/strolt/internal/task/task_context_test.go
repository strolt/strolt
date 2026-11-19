package task

import (
	"testing"
	time "time"

	config "github.com/strolt/strolt/apps/strolt/internal/config"

	stdcontext "context"

	"github.com/stretchr/testify/require"

	taskcontext "github.com/strolt/strolt/apps/strolt/internal/context"
)

func TestClonePreservesProcessContext(t *testing.T) {
	processContext, cancel := stdcontext.WithCancel(stdcontext.Background())
	defer cancel()

	original := Task{Context: taskcontext.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}, Trigger: "", ServiceName: "", TaskName: "", OperationBackup: Operation{Schedule: "", LatestRun: 0, LatestDuration: 0, IsPreviouslyLaunched: false}, OperationPrune: Operation{Schedule: "", LatestRun: 0, LatestDuration: 0, IsPreviouslyLaunched: false}, TaskConfig: config.TaskConfig{Schedule: config.Schedule{Backup: "", Prune: ""}, Source: config.DriverSourceConfig{Driver: "", Config: nil, Env: nil}, OperationMode: "", Destinations: nil, Notifications: nil, Tags: nil}, log: nil, isNotificationsDisabled: false, notificationWaitGroup: nil}
	original.Context.SetProcessContext(processContext)
	cloned, err := original.Clone()
	require.NoError(t, err)
	require.Same(t, processContext, cloned.Context.ProcessContext())

	cancel()
	require.ErrorIs(t, cloned.Context.ProcessContext().Err(), stdcontext.Canceled)
}

func TestProcessContextCancellationIsVisible(t *testing.T) {
	processContext, cancel := stdcontext.WithCancel(stdcontext.Background())
	operationContext := taskcontext.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}
	operationContext.SetProcessContext(processContext)

	cancel()
	require.ErrorIs(t, operationContext.ProcessContext().Err(), stdcontext.Canceled)
}
