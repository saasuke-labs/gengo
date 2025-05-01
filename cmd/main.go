package main

import (
	"blog-down/pkg/cli"
	"blog-down/pkg/generator"
	"blog-down/pkg/server"
	"blog-down/pkg/watcher"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	_ "github.com/yuin/goldmark/extension"
)

var rootCmd cobra.Command

func execGenerateSite(postsPath, outputPath string, watchMode bool) {
	ch, filesProgress := generator.GenerateSiteAsync(postsPath, outputPath, watchMode)

	files := []string{}
	filesStatuses := map[string]generator.FileStatus{}
	completed := 0

	for _, fileProgress := range filesProgress {
		files = append(files, fileProgress.Filename)
		filesStatuses[fileProgress.Filename] = fileProgress.Status
	}

	cli.UpdateScreen("Files progress:", files, filesStatuses, completed, len(files))

	fmt.Println("Waiting for updates...")
	for {
		select {
		case progress, ok := <-ch:
			if !ok {
				// Channel closed, processing complete
				fmt.Println("Channel closed")
				return
			}

			fmt.Println("Received update:", progress.Filename, progress.Status)
			// Update our state
			filesStatuses[progress.Filename] = progress.Status
			if progress.Status == generator.Completed || progress.Status == generator.Failed {
				completed++
			}
			cli.UpdateScreen("Files progress:", files, filesStatuses, completed, len(files))

		default:
			// No updates available, wait a bit before refreshing
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func execServeSite(sitePath string, watchMode bool, port int) {
	server.Serve(sitePath, watchMode, port)
}

func init() {

	rootCmd = cobra.Command{
		Use:   "rego",
		Short: "Static site generator from Markdown",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	var postsPath string
	var sitePath string
	var outputPath string
	var watchMode bool

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate the static site",
		Long:  `Generate the static site from the manifest.yaml file and output it to the specified directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			// This function will be executed when the "subcommand" is called
			//execGenerateSite(postsPath, outputPath, watchMode)

			// TODO - See how to find this directory from posts.yaml
			go watcher.WatchDir("./blog/posts", func(file string) {
				// TODO - Optimize and generate only the changed file
				fmt.Println(("Generating site..."))
				execGenerateSite(postsPath, outputPath, watchMode)

			})

			fmt.Println("Press any key to exit...")
			var b []byte = make([]byte, 1)
			os.Stdin.Read(b)
		},
	}

	generateCmd.Flags().StringVar(&postsPath, "posts", "posts.yaml", "Path to posts.yaml")
	generateCmd.Flags().StringVar(&outputPath, "output", "output", "Output directory")
	generateCmd.Flags().BoolVar(&watchMode, "watch", false, "Enable watch mode with hot reload")

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serves the static site",
		Long:  `Serves the static site.`,
		Run: func(cmd *cobra.Command, args []string) {
			// This function will be executed when the "subcommand" is called
			//TODO -  Get port from flags
			execServeSite(sitePath, watchMode, 3000)
		},
	}

	serveCmd.Flags().StringVar(&sitePath, "site", "site", "Site directory")
	serveCmd.Flags().BoolVar(&watchMode, "watch", false, "Enable watch mode with hot reload")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(serveCmd)
}

func main() {
	rootCmd.Execute()

	// files := []string{"file1.md", "file2.md", "file3.md"}
	// statuses := []cli.FileStatus{cli.Started, cli.Completed}

	// fileStatuses := map[string]cli.FileStatus{}
	// for _, file := range files {
	// 	fileStatuses[file] = cli.Pending
	// }

	// i := 0
	// status := 0
	// completed := 0

	// for {
	// 	status = (status + 1) % len(statuses)

	// 	if status == 0 {
	// 		i++
	// 	}

	// 	if status == len(statuses)-1 {
	// 		completed++
	// 	}
	// 	fileStatuses[files[i]] = statuses[status]

	// 	cli.UpdateScreen("This is a test", files, fileStatuses, completed, len(files))
	// 	time.Sleep(1 * time.Second)

	// }

}

// func watchAndRebuild(postsPath, outputPath string) {
// 	w, _ := fsnotify.NewWatcher()
// 	defer w.Close()
// 	//contentDir := filepath.Dir(postsPath)
// 	w.Add(postsPath)
// 	w.Add("templates")
// 	filepath.WalkDir("content", func(path string, d fs.DirEntry, err error) error {
// 		if !d.IsDir() && strings.HasSuffix(path, ".md") {
// 			w.Add(path)
// 		}
// 		return nil
// 	})
// 	for {
// 		select {
// 		case <-w.Events:
// 			generator.GenerateSite(postsPath, outputPath, true)
// 			broadcastReload()
// 		case err := <-w.Errors:
// 			log.Println("watch error:", err)
// 		}
// 	}
// }

// func serveOutput(outputPath string) {
// 	http.Handle("/", http.FileServer(http.Dir(outputPath)))
// 	http.HandleFunc("/reload", handleSSE)
// 	log.Println("Serving on http://localhost:8080")
// 	http.ListenAndServe(":8080", nil)
// }

// func handleSSE(w http.ResponseWriter, r *http.Request) {
// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	client := make(chan string)
// 	sseRegister <- client
// 	defer func() { sseUnregister <- client }()
// 	for {
// 		select {
// 		case <-client:
// 			fmt.Fprintf(w, "data: reload\n\n")
// 			flusher.Flush()
// 		case <-r.Context().Done():
// 			return
// 		}
// 	}
// }

// func startSSEServer() {
// 	for {
// 		select {
// 		case c := <-sseRegister:
// 			sseClients[c] = true
// 		case c := <-sseUnregister:
// 			delete(sseClients, c)
// 			close(c)
// 		}
// 	}
// }

// func broadcastReload() {
// 	for c := range sseClients {
// 		c <- "reload"
// 	}
// }
