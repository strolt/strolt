package config

import (
	"context"
	"encoding/json"

	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"

	"github.com/fsnotify/fsnotify"
)

func WatchConfigChanges(ctx context.Context, cancel func()) {
	if env.IsWatchFilesDisabled() {
		return
	}

	log := logger.New()

	fileListByte, err := json.Marshal(fileList)
	if err != nil {
		log.Error(err)
	}

	log.WithField("files", string(fileListByte)).Info("start watch for config file changes")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		<-ctx.Done()

		log.Debug("stop watch for config file changes...")
		done <- true
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Infof("file '%s' changed", event.Name)
				cancel()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Errorf("config file watcher: %s", err)
			}
		}
	}()

	for _, filepath := range fileList {
		if err := watcher.Add(filepath); err != nil {
			log.Error(err)
		}
	}

	<-done
	log.Debug("watch for config file changes stopped")
}
