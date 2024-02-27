package assets

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watcher provides a file watching mechanism for the
// assets folder. It will watch for changes in the assets
// and copy them to the destination folder as these change
// to keep in sync the assets folder with the destination
// folder.
// At the beginning of the program, the assets folder is
// copied to the destination folder, and then the watcher
// is started.
func Watcher(src, dest string) func() {
	return func() {
		err := CopyToPublic(src, dest)
		if err != nil {
			log.Println(err)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(fmt.Errorf("error creating watcher: %w", err))
		}

		// Add all folders within the assets folder to the watcher.
		err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			return watcher.Add(path)
		})

		if err != nil {
			panic(fmt.Errorf("error adding files to watcher: %w", err))
		}

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						continue
					}

					needsCopy := event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Rename)
					if !needsCopy {
						continue
					}

					err := CopyToPublic(src, dest)
					if err != nil {
						log.Println(err)
					}

					if event.Has(fsnotify.Create) {
						watcher.Add(event.Name)
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}

					log.Println("error:", err)
				}
			}
		}()

		<-make(chan struct{})
	}
}
