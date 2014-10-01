package server

import (
	//"fmt"
	"github.com/rcrowley/go-tigertonic"
	//"net/http"
)

// Start a HTTP listener
func Start() {
	mux := tigertonic.NewTrieServeMux()
	mux.Handle("GET", "/api/blocks", tigertonic.Timed(tigertonic.Marshaled(GetHello), "GetHelloHandler", nil))
	mux.Handle("GET", "/api/blocks/download/{itemID}", tigertonic.Timed(NewFileDownloadHandler(), "FileDownloadHandler", nil))
	mux.Handle("POST", "/api/blocks/upload", tigertonic.Timed(NewPostMultipartUploadHandler(), "PostMultipartUploadHandler", nil))
	mux.Handle("PUT", "/api/blocks/upload", tigertonic.Timed(NewRawUploadHandler(), "RawUploadHandler", nil))
	tigertonic.NewServer(":8002", mux).ListenAndServe()

	/*mux.Handle("GET", "/api/blocks/spaces", tigertonic.Timed(tigertonic.Marshaled(handlers.GetStorageSpaces), "GetSpacesHandler", nil))
	mux.Handle("GET", "/storage/metadata/item/{spaceID}/{itemID}", tigertonic.Timed(tigertonic.Marshaled(handlers.GetItem), "GetItemHandler", nil))
	mux.Handle("GET", "/storage/metadata/item-children/{spaceID}/{itemID}", tigertonic.Timed(tigertonic.Marshaled(handlers.GetItemChildren), "GetItemChildrenHandler", nil))*/

	//mux.Handle("POST", "/api/blocks/download/{spaceID}/{folderID}", tigertonic.Timed(handlers.NewMultiPartFileUploadHandler(), "PostMultipartStorageFileHandler", nil))
	// mux.Handle("PUT", "/storage/content/upload/{spaceID}/{folderID}", tigertonic.Timed(handlers.NewRawFileUploadHandler(), "RawFileUploadHandler", nil))
	// tigertonic.NewServer(":8002", tigertonic.Logged(mux, nil)).ListenAndServe()

	/*mux := http.NewServeMux()
	mux.HandleFunc("/api/block", blockHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to Blocker!")
	})

	http.ListenAndServe(":8001", mux)*/
}
