package main

import (
	"log"
	"os"

	"github.com/manics/aws-ecr-registry-cleaner/amazon"
)

var (
	// Version is set at build time using the Git repository metadata
	Version string
)

// The main entrypoint for the service
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		log.Println(Version)
		os.Exit(0)
	}

	ecrH, err := amazon.Setup(os.Args[1:])
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	errs := ecrH.RunOnce()
	if len(errs) > 0 {
		log.Fatalf("ERROR: %s", errs)
	}
}
