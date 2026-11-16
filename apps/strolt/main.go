// Strolt is a CLI and service for managing, scheduling, and monitoring backups.
package main

import (
	"context"
	"os/signal"
	"syscall"

	cmd "github.com/strolt/strolt/apps/strolt/cmd/strolt"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cmd.ExecuteContext(ctx)
}
