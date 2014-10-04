package server

import (
	//"fmt"
	"github.com/rcrowley/go-tigertonic"
	//"net/http"
)

// Start a HTTP listener
func Start() {
	mux := tigertonic.NewTrieServeMux()
	mux.Handle("GET", "/api/blocker", tigertonic.Timed(tigertonic.Marshaled(GetHello), "GetHelloHandler", nil))
	mux.Handle("GET", "/api/blocker/{itemID}/{fileName}", tigertonic.Timed(NewFileDownloadHandler(), "FileDownloadHandler", nil))
	mux.Handle("POST", "/api/blocker", tigertonic.Timed(NewPostMultipartUploadHandler(), "PostMultipartUploadHandler", nil))
	mux.Handle("PUT", "/api/blocker", tigertonic.Timed(NewRawUploadHandler(), "RawUploadHandler", nil))
	tigertonic.NewServer(":8002", mux).ListenAndServe()
}
