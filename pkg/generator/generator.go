package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type FileStatus string

const (
	Pending   FileStatus = "pending"
	Started   FileStatus = "started"
	Completed FileStatus = "completed"
	Failed    FileStatus = "failed"
)

// FileProgress represents a progress update for a file
type FileProgress struct {
	Filename string
	Status   FileStatus
}

func GenerateSiteAsync(manifestPaths []string, outputDir string) ([]FileProgress, <-chan FileProgress) {

	manifest := getManifest(manifestPaths)

	// TODO - See this
	baseDir := filepath.Dir(manifestPaths[0])

	fmt.Println("Generating site...", manifest)
	progressCh := make(chan FileProgress)

	tasks := scheduleTasks(manifest, baseDir, outputDir)

	files := make([]FileProgress, len(tasks))
	for idx, task := range tasks {

		files[idx] = FileProgress{
			Filename: task.Name(),
			Status:   Pending,
		}
	}

	go func() {
		var wg sync.WaitGroup

		for _, task := range tasks {
			wg.Add(1)
			go func(task Task) {
				defer wg.Done()

				progressCh <- FileProgress{Filename: task.Name(), Status: Started}

				err := task.Execute()

				if err == nil {
					progressCh <- FileProgress{Filename: task.Name(), Status: Completed}
				} else {
					progressCh <- FileProgress{Filename: task.Name(), Status: Failed}
				}

			}(task)
		}

		wg.Wait()
		close(progressCh)
	}()

	return files, progressCh
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}
