package main

import (
	"flag"
	"fmt"
	"github.com/keithballdotnet/blocker/blocks"
	"github.com/keithballdotnet/blocker/server"
	"log"
	"os"
	"strings"
)

const AppVersion = "1.0.4"

func main() {

	// Set up executable flags
	version := flag.Bool("v", false, "prints current version without starting the application")
	storageProvider := flag.String("s", "nfs", "Storage provider selection either 'nfs', 'cb', 'azure' or 's3'")
	cryptoProvider := flag.String("c", "openpgp", "Crypto provider selection either 'gokms', 'openpgp' or 'aws'")

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

	// Ensure string is to lower
	blocks.CryptoProviderName = strings.ToLower(*cryptoProvider)

	// Validate storage provider
	if blocks.CryptoProviderName != "openpgp" && blocks.CryptoProviderName != "aws" && blocks.CryptoProviderName != "gokms" {
		fmt.Println("Unknown Provider: Crypto provider selection either 'gokms', 'openpgp' or 'aws'")
		os.Exit(0)
	}

	// Now set up repos
	blocks.SetUpRepositories()

	log.SetOutput(os.Stdout)
	log.SetPrefix("Blocker:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Starting Blocker: %s - Using Provider: %s", AppVersion, blocks.StorageProviderName)

	// Start the server
	server.Start()
}
