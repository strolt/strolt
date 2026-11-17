// Strolt is a CLI and service for managing, scheduling, and monitoring backups.
package main

import (
	"context"
	"os/signal"
	"syscall"

	cmd "github.com/strolt/strolt/apps/strolt/cmd/strolt"
	"github.com/strolt/strolt/shared/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := cmd.ExecuteContext(ctx); err != nil {
		logger.New().Fatal(err)
	}
}
