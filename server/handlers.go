package server

import (
	"fmt"
	"net/http"
)

func blockHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Welcome to the block handler")
}
