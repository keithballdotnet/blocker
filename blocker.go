package main

import (
	"flag"
	"fmt"
	"github.com/Inflatablewoman/blocker/server"
	"log"
	"os"
)

const AppVersion = "1.0.1 beta"

func main() {

	// This code allows someone to ask what version I am from the command line
	version := flag.Bool("v", false, "prints current version without starting the application")
	flag.Parse()
	if *version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	log.SetOutput(os.Stdout)
	log.SetPrefix("Blocker:")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting Blocker: ", AppVersion)

	// Start the server
	server.Start()
}
