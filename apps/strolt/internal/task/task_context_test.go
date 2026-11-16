package task

import (
	stdcontext "context"
	"testing"

	"github.com/stretchr/testify/require"
	taskcontext "github.com/strolt/strolt/apps/strolt/internal/context"
)

func TestClonePreservesProcessContext(t *testing.T) {
	processContext, cancel := stdcontext.WithCancel(stdcontext.Background())
	defer cancel()

	original := Task{}
	original.Context.SetProcessContext(processContext)
	cloned, err := original.Clone()
	require.NoError(t, err)
	require.Same(t, processContext, cloned.Context.ProcessContext())

	cancel()
	require.ErrorIs(t, cloned.Context.ProcessContext().Err(), stdcontext.Canceled)
}

func TestProcessContextCancellationIsVisible(t *testing.T) {
	processContext, cancel := stdcontext.WithCancel(stdcontext.Background())
	operationContext := taskcontext.Context{}
	operationContext.SetProcessContext(processContext)

	cancel()
	require.ErrorIs(t, operationContext.ProcessContext().Err(), stdcontext.Canceled)
}
