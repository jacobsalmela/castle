package config

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func WatchConfig(cfgPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create config watcher: %v (config hot-reload disabled)", err)
		return
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write != 0 && event.Name == cfgPath {
					// Use merge to handle new/deprecated fields
					cfg, mergeResult, err := LoadConfigWithMerge(cfgPath)
					if err != nil {
						log.Println("config reload error:", err)
						continue
					}
					Cfg = cfg

					// Log merge information
					if mergeResult != nil {
						if len(mergeResult.NewFields) > 0 {
							log.Printf("Added %d new config fields with defaults", len(mergeResult.NewFields))
						}
						if len(mergeResult.DeprecatedFields) > 0 {
							log.Printf("Warning: %d deprecated fields found (will be ignored)", len(mergeResult.DeprecatedFields))
						}
					}

					// Note: Window size is NOT changed on hot-reload - user controls window size
					log.Println("(Hot) reloaded config:", cfgPath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()

	// Check if file exists before trying to watch it
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Printf("File %s does not exist yet, skipping watch", cfgPath)
		return
	}

	if err := watcher.Add(cfgPath); err != nil {
		log.Printf("Failed to watch config file %s: %v (config hot-reload disabled)", cfgPath, err)
		return
	}
	select {} // block forever
}
