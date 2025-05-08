package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/tonitienda/gengo/pkg/generator"
	"github.com/tonitienda/gengo/pkg/watcher"
)

func generate(manifestPath, outputPath string) {
	files, ch := generator.GenerateSiteAsync(manifestPath, outputPath)

	filesStatuses := make(map[string]generator.FileStatus)
	fileNames := make([]string, len(files))

	for i, file := range files {
		fileNames[i] = file.Filename
		filesStatuses[file.Filename] = file.Status
	}

	completed := 0

	//fmt.Println("Waiting for updates...")
	for {
		select {
		case progress, ok := <-ch:
			if !ok {
				fmt.Printf("Generated %d / %d files\n", completed, len(files))
				return
			}

			if progress.Status == generator.Completed || progress.Status == generator.Failed {
				completed++
			}

			filesStatuses[progress.Filename] = progress.Status

			UpdateScreen("Files progress:", fileNames, filesStatuses, completed, len(files))

		default:
			// No updates available, wait a bit before refreshing
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func Generate(manifestPath, outputPath string, watchMode bool) {
	generate(manifestPath, outputPath)

	if watchMode {
		// TODO - See how to find this directory from posts.yaml
		go watcher.WatchDir("./blog", func(file string) {
			// TODO - Optimize and generate only the changed file
			//fmt.Println(("Generating site..."))
			generate(manifestPath, outputPath)

		})

		//fmt.Println("Press any key to exit...")
		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)
	}
}

func SilentGenerate(manifestPath, outputPath string) {
	files, ch := generator.GenerateSiteAsync(manifestPath, outputPath)

	completed := 0

	//fmt.Println("Waiting for updates...")
	for {
		select {
		case progress, ok := <-ch:
			if !ok {
				fmt.Printf("Generated %d / %d files\n", completed, len(files))
				return
			}

			if progress.Status == generator.Completed || progress.Status == generator.Failed {
				completed++
				fmt.Printf("File %s: %s\n", progress.Filename, progress.Status)
			}

		default:
			// No updates available, wait a bit before refreshing
			time.Sleep(100 * time.Millisecond)
		}
	}

}
