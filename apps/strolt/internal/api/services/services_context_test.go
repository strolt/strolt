package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOperationContextPreservesRequestValuesAndDeadline(t *testing.T) {
	type contextKey struct{}
	value := "request-value"
	deadline := time.Now().Add(time.Minute)
	requestContext, requestCancel := context.WithDeadline(context.WithValue(context.Background(), contextKey{}, value), deadline)
	defer requestCancel()

	operationContext, cleanup := New(context.Background(), nil).operationContext(requestContext)
	defer cleanup()

	gotDeadline, ok := operationContext.Deadline()
	require.True(t, ok)
	require.Equal(t, deadline, gotDeadline)
	require.Equal(t, value, operationContext.Value(contextKey{}))
}

func TestOperationContextCanceledRequest(t *testing.T) {
	requestContext, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	operationContext, cleanup := New(context.Background(), nil).operationContext(requestContext)
	defer cleanup()

	require.ErrorIs(t, operationContext.Err(), context.Canceled)
}

func TestOperationContextCanceledDaemon(t *testing.T) {
	cause := errors.New("daemon stopped")
	daemonContext, cancelDaemon := context.WithCancelCause(context.Background())
	cancelDaemon(cause)

	operationContext, cleanup := New(daemonContext, nil).operationContext(context.Background())
	defer cleanup()

	require.ErrorIs(t, operationContext.Err(), context.Canceled)
	require.ErrorIs(t, context.Cause(operationContext), cause)
}
