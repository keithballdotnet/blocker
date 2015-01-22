package server

import (
	// "fmt"
	//"github.com/Inflatablewoman/blocker/crypto"
	"github.com/rcrowley/go-tigertonic"
)

// Start a HTTP listener
func Start() {
	mux := tigertonic.NewTrieServeMux()
	mux.Handle("GET", "/api/blocker", tigertonic.Timed(tigertonic.Marshaled(GetHello), "GetHelloHandler", nil))
	mux.Handle("GET", "/api/blocker/{itemID}", tigertonic.Timed(NewFileDownloadHandler(), "FileDownloadHandler", nil))
	mux.Handle("DELETE", "/api/blocker/{itemID}", tigertonic.Timed(tigertonic.Marshaled(DeleteHandler), "DeleteHandler", nil))
	mux.Handle("COPY", "/api/blocker/{itemID}", tigertonic.Timed(tigertonic.Marshaled(CopyHandler), "CopyHandler", nil))
	mux.Handle("POST", "/api/blocker", tigertonic.Timed(NewPostMultipartUploadHandler(), "PostMultipartUploadHandler", nil))
	mux.Handle("PUT", "/api/blocker", tigertonic.Timed(NewRawUploadHandler(), "RawUploadHandler", nil))
	// Log to Console
	tigertonic.NewServer(":8010", tigertonic.ApacheLogged(mux)).ListenAndServe()
	//tigertonic.NewServer(":8002", mux).ListenAndServe()

	// Inititin the keypath will be enough to create the certificates if needed
	//tigertonic.NewServer(":8002", mux).ListenAndServeTLS(crypto.RsaEncryptionChipher.PrivateKeyPath, crypto.RsaEncryptionChipher.PublicKeyPath)
}
