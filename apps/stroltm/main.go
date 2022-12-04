package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/strolt/strolt/apps/stroltm/internal/api"
	"github.com/strolt/strolt/apps/stroltm/internal/config"
	"github.com/strolt/strolt/apps/stroltm/internal/env"
	"github.com/strolt/strolt/apps/stroltm/internal/logger"
	"github.com/strolt/strolt/apps/stroltm/internal/manager"
)

func main() {
	env.Scan()

	log := logger.New()

	if err := config.Scan(); err != nil {
		log.Error(err)
		return
	}

	log.Info(config.Get())

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	c := make(chan os.Signal, 1)
	defer close(c)

	signal.Notify(c, os.Interrupt)

	{ // Api server
		wg.Add(1)
		go func() {
			api.New().Run(ctx, cancel)
			wg.Done()
		}()
	}

	{ // Manager
		wg.Add(1)
		go func() {
			manager.New().Watch(ctx, cancel)
			wg.Done()
		}()
	}

	// Watch system exit code
	go func() {
		oscall := <-c
		log.Debugf("system call: %+v", oscall)
		cancel()
	}()

	wg.Wait()
}
