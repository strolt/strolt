package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/logger"

	"github.com/fsnotify/fsnotify"
)

func WatchConfigChanges(ctx context.Context, cancel func()) {
	if !config.DisableWatchChanges {
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

			log.Info("stop watch for config file changes")
			done <- true
		}()

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					log.Info(fmt.Sprintf("file '%s' changed", event.Name))
					cancel()

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}

					log.Error(fmt.Sprintf("config file watcher: %s", err))
				}
			}
		}()

		for _, filepath := range fileList {
			if err := watcher.Add(filepath); err != nil {
				log.Error(err)
			}
		}

		<-done
		log.Info("watch for config file changes stopped")
	}
}
