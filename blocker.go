package main

import (
	"flag"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	"github.com/Inflatablewoman/blocker/crypto"
	"github.com/Inflatablewoman/blocker/server"
	"log"
	"os"
	"strings"
)

const AppVersion = "1.0.4"

func main() {

	// Set up executable flags
	version := flag.Bool("v", false, "prints current version without starting the application")
	storageProvider := flag.String("s", "nfs", "Storage provider selection either 'nfs', 'cb', 'azure' or 's3'")

	// This code allows someone to ask what version I am from the command line

	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	// Ensure string is to lower
	blocks.StorageProviderName = strings.ToLower(*storageProvider)

	// Validate storage provider
	if blocks.StorageProviderName != "nfs" && blocks.StorageProviderName != "cb" && blocks.StorageProviderName != "azure" && blocks.StorageProviderName != "s3" {
		fmt.Println("Unknown Provider: Storage provider selection either 'nfs', 'cb', 'azure' or 's3'")
		os.Exit(0)
	}

	// Now set up repos
	blocks.SetUpRepositories()

	// Get the PGP keys
	crypto.GetPGPKeyRings()

	log.SetOutput(os.Stdout)
	log.SetPrefix("Blocker:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Starting Blocker: %s - Using Provider: %s", AppVersion, blocks.StorageProviderName)

	// Start the server
	server.Start()
}
