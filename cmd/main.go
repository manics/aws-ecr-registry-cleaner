package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/manics/aws-ecr-registry-cleaner/amazon"
)

var (
	// Version is set at build time using the Git repository metadata
	Version string
)

// The main entrypoint for the service
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	versionFlag := flag.Bool("version", false, "Display the version and exit")
	loopDelayFlag := flag.Int("loop-delay", 0, "Run the service in a loop, sleep for this many seconds between runs, default 0=run once")
	dryRunFlag := flag.Bool("dry-run", false, "Dry run, do not delete anything")
	awsRegistryIdFlag := flag.String("aws-registry-id", "", "AWS registry ID, default account ID of the credentials")
	expiresAfterPullDays := flag.Int("expires-after-pull-days", 7, "Delete images that have not been pulled in this many days, set to 0 to delete all images")

	flag.Parse()

	if *versionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *expiresAfterPullDays < 0 {
		log.Fatalf("expiresAfterPullDays must be >= 0")
	}

	log.Printf("Version: %s", Version)
	ecrH, err := amazon.Setup(*dryRunFlag, *awsRegistryIdFlag, *expiresAfterPullDays)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	loopEnabled := *loopDelayFlag > 0
	for {
		errs := ecrH.RunOnce()
		if len(errs) > 0 {
			if loopEnabled {
				log.Printf("ERROR: %s", errs)
			} else {
				log.Fatalf("ERROR: %s", errs)
			}
		}
		if loopEnabled {
			time.Sleep(time.Duration(*loopDelayFlag) * time.Second)
		} else {
			break
		}
	}
}
