package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/saasuke-labs/gengo/pkg/generator"
	"github.com/saasuke-labs/gengo/pkg/telemetry"
	"github.com/saasuke-labs/gengo/pkg/watcher"
	"github.com/spf13/cobra"
)

func NewGenerateCommand() *cobra.Command {

	var manifestPath string
	var outputPath string
	var watchMode bool
	var plainMode bool

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate the static site",
		Long:  `Generate the static site from the manifest.yaml file and output it to the specified directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			defer telemetry.Close()
			telemetry.Track("generate-started", map[string]interface{}{
				"command": "generate",
				"plain":   plainMode,
			})
			if plainMode {
				SilentGenerate(manifestPath, outputPath)
				return
			}
			Generate(manifestPath, outputPath, watchMode)
			telemetry.Track("generate-completed", map[string]interface{}{
				"command": "generate",
				"plain":   plainMode,
			})
		},
	}

	generateCmd.Flags().StringVar(&manifestPath, "manifest", "gengo.yaml", "Path to the manifest file")
	generateCmd.Flags().StringVar(&outputPath, "output", "output", "Output directory")
	generateCmd.Flags().BoolVar(&watchMode, "watch", false, "Enable watch mode with hot reload")
	generateCmd.Flags().BoolVar(&plainMode, "plain", false, "Plain output. Useful for non-interactive shell")

	return generateCmd
}

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
