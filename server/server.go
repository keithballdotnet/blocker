package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	// "fmt"
	"flag"
	"github.com/Inflatablewoman/blocker/crypto"
	"log"

	"github.com/rcrowley/go-tigertonic"
)

var (
	// cert path
	cert = flag.String("cert", "", "SSL Certificate path")
	// cert key path
	certKey = flag.String("certkey", "", "SSL Private key path")
	// Shared key path
	sharedKeyPath = flag.String("sharedKey", "", "Shared Authentication Key path")
)

var (
	// This key is used for authentication with the server
	SharedKey = ""
)

// Start a HTTP listener
func Start() {

	// Set up the auth key
	SetupAuthenticationKey()

	// Set-up API listeners
	mux := tigertonic.NewTrieServeMux()
	mux.Handle("GET", "/api/v1/blocker", tigertonic.Timed(tigertonic.Marshaled(GetHello), "GetHelloHandler", nil))
	mux.Handle("GET", "/api/v1/blocker/{itemID}", tigertonic.Timed(NewFileDownloadHandler(), "FileDownloadHandler", nil))
	mux.Handle("DELETE", "/api/v1/blocker/{itemID}", tigertonic.Timed(tigertonic.Marshaled(DeleteHandler), "DeleteHandler", nil))
	mux.Handle("COPY", "/api/v1/blocker/{itemID}", tigertonic.Timed(tigertonic.Marshaled(CopyHandler), "CopyHandler", nil))
	mux.Handle("POST", "/api/v1/blocker", tigertonic.Timed(NewPostMultipartUploadHandler(), "PostMultipartUploadHandler", nil))
	mux.Handle("PUT", "/api/v1/blocker", tigertonic.Timed(NewRawUploadHandler(), "RawUploadHandler", nil))
	// Log to Console
	server := tigertonic.NewServer(":8010", tigertonic.ApacheLogged(mux))
	if *certKey == "" || *cert == "" {
		server.ListenAndServe()
	} else {
		log.Println("SSL Enabled")
		if err := server.ListenAndServeTLS(*cert, *certKey); err != nil {
			log.Fatal(err)
		}
	}
}

// SetupAuthenticationKey  - This deals with setting an auth key for the service
func SetupAuthenticationKey() {

	// Locate key file
	sharedKeyFromCLI := true
	keyPath := ""
	if *sharedKeyPath == "" {
		sharedKeyFromCLI = false
		defaultAuthDir := filepath.Join(os.TempDir(), "blocker")
		err := os.Mkdir(defaultAuthDir, 0777)
		if err != nil && !os.IsExist(err) {
			panic("Unable to create directory: " + err.Error())
		}

		keyPath = filepath.Join(defaultAuthDir, "auth.key")
	} else {
		keyPath = *sharedKeyPath
	}

	// Read the auth key file
	bytes, err := ioutil.ReadFile(keyPath)

	// The key file could not be located and is expected.  Get out of here...
	if err != nil && os.IsNotExist(err) && sharedKeyFromCLI {
		panic("Unable to read shared key file: " + err.Error())
	}

	// No file present.  Let's create a key.
	if err != nil && os.IsNotExist(err) {
		newAccessKey := strings.ToLower(crypto.RandomSecret(40))
		log.Printf("Generated new Access Key: %s", newAccessKey)
		// Write key to key file
		err := ioutil.WriteFile(keyPath, []byte(newAccessKey), 0644)
		if err != nil {
			panic("Unable to write shared key file: " + err.Error())
		}

		// Set the key
		SharedKey = newAccessKey

		// We're done.
		return
	}

	log.Printf("Using auth key file: %v", keyPath)

	// Get the key
	SharedKey = string(bytes)
}
