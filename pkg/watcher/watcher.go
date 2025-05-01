package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func WatchDir(dirPath string, callback func(string)) {
	fmt.Println("Adding watcher for directory:", dirPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fmt.Println("Adding watcher for directory:", path)
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			log.Println("File changed:", event.Name)
			//if strings.HasSuffix(event.Name, ".mdx") {
			fmt.Println("Calling callback for file:", event.Name)
			callback(event.Name)
			//}
		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}
