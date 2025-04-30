package server

import (
	"fmt"
	"net/http"
)

func Serve(sitePath string, watchMode bool) {
	fmt.Println("Serving site at", sitePath, "from http://localhost:3000")
	http.Handle("/", http.FileServer(http.Dir(sitePath)))
	http.ListenAndServe(":3000", nil)
}
