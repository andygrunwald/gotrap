// gotrap - A Gerrit <=> Github <=> TravisCI bridge
package main

import (
	"flag"
	"fmt"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/stream"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var (
	flagConfigFile *string
	flagPidFile    *string
	flagVersion    *bool
)

const (
	majorVersion = 1
	minorVersion = 0
	patchVersion = 0
)

func init() {
	flagConfigFile = flag.String("config", "", "Path to configuration file.")
	flagPidFile = flag.String("pidfile", "", "Write the process id into a given file.")
	flagVersion = flag.Bool("version", false, "Outputs the version number and exits.")
}

// main is the heart of gotrap.
func main() {
	flag.Parse()

	// Output the version and exit
	if *flagVersion {
		fmt.Printf("gotrap v%d.%d.%d\n", majorVersion, minorVersion, patchVersion)
		return
	}

	// Check for configuration file
	if len(*flagConfigFile) <= 0 {
		log.Fatal("No configuration file found. Please add the --config parameter")
	}

	// PID-File
	if len(*flagPidFile) > 0 {
		ioutil.WriteFile(*flagPidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	}

	// Be nice to the user
	log.Println("Hey, nice to meet you. Just wait a second. I will start up.")
	defer log.Println("Our job is done. We have to go.")

	// Bootstrap configuration file
	config, err := config.NewConfiguration(flagConfigFile)
	if err != nil {
		log.Fatal("Configuration initialisation failed:", err)
	}

	// Bootstrap stream
	stream, err := stream.GetStream(stream.StreamAmqp)
	if err != nil {
		log.Fatal("Stream initialisation failed:", err)
	}

	stream.Initialize(config)
	err = stream.Start()
	if err != nil {
		log.Fatal("Stream start failed:", err)
	}
}
