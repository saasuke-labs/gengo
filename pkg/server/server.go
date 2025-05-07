package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tonitienda/gengo/pkg/watcher"

	"github.com/gorilla/websocket"
)

func fileHandler(sitePath string, watchMode bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(sitePath, r.URL.Path)
		info, err := os.Stat(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.IsDir() {
			path = filepath.Join(path, "index.html")
		}

		if watchMode && strings.HasSuffix(path, ".html") {
			content, err := os.ReadFile(path)
			if err != nil {
				http.Error(w, "Failed to read file", 500)
				return
			}
			// Inject the WebSocket script before closing </body>
			html := string(content)
			injected := strings.Replace(html, "</body>", `
				<script>
					const ws = new WebSocket("ws://localhost:3000/ws");
					ws.onmessage = () => window.location.reload();
				</script>
			</body>`, 1)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(injected))
		} else {
			http.ServeFile(w, r, path)
		}
	})
}

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()
	clients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}
	}
}

// TODO - Notify clients only when the file
// that affects the open page changed
func notifyClients(filePath string) {
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte("reload"))
		if err != nil {
			log.Println("WebSocket write error:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

func Serve(sitePath string, watchMode bool, port int) {
	fmt.Println("Serving site at", sitePath, "from http://localhost:", port)

	http.Handle("/", fileHandler(sitePath, watchMode))
	if watchMode {
		http.HandleFunc("/ws", wsHandler)
		go watcher.WatchDir(sitePath, notifyClients)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
