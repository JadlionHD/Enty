// Package configwatch provides utility functions for config file watching without external dependencies.
package configwatch

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// WatchConfigFile watches a config file for changes using polling (mod time).
// onChange is called whenever the file is modified.
func WatchConfigFile(configPath string, interval time.Duration, onChange func()) (stop func()) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Printf("WatchConfigFile: %v", err)
		return func() {}
	}
	stopChan := make(chan struct{})
	go func() {
		var lastModTime time.Time
		for {
			select {
			case <-stopChan:
				return
			default:
				info, err := os.Stat(absPath)
				if err == nil {
					modTime := info.ModTime()
					if modTime.After(lastModTime) {
						onChange()
						lastModTime = modTime
					}
				}
				time.Sleep(interval)
			}
		}
	}()
	return func() { close(stopChan) }
}
