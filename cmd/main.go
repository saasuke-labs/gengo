package main

import (
	"github.com/saasuke-labs/gengo/pkg/cli"
	"github.com/saasuke-labs/gengo/pkg/server"

	"github.com/spf13/cobra"
	_ "github.com/yuin/goldmark/extension"
)

var rootCmd cobra.Command

func execServeSite(sitePath string, watchMode bool, port int) {
	server.Serve(sitePath, watchMode, port)
}

func init() {

	var sitePath string
	var watchMode bool

	rootCmd = cobra.Command{
		Use:   "gengo",
		Short: "Static site generator from Markdown",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

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

	rootCmd.AddCommand(cli.NewGenerateCommand())
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(cli.NewVersionCommand())
}

func main() {
	rootCmd.Execute()

}

// func watchAndRebuild(manifestPath, outputPath string) {
// 	w, _ := fsnotify.NewWatcher()
// 	defer w.Close()
// 	//contentDir := filepath.Dir(manifestPath)
// 	w.Add(manifestPath)
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
// 			generator.GenerateSite(manifestPath, outputPath, true)
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
