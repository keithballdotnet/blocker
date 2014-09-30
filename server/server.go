package server

import (
	"fmt"
	"net/http"
)

// Start a HTTP listener
func Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/block", blockHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to Blocker!")
	})

	http.ListenAndServe(":8001", mux)
}
