package main

import (
	"flag"
	"fmt"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/stream"
	"log"
)

// Global variable to store the configuration file
var (
	flagConfigFile *string
	flagVersion    *bool
)

const (
	MajorVersion = 1
	MinorVersion = 0
	PatchVersion = 0
)

// Init function to define arguments
func init() {
	flagConfigFile = flag.String("config", "./config.json", "Configuration file")
	flagVersion = flag.Bool("version", false, "Outputs the version number and exits")
}

// The heart of gotrap.
func main() {
	flag.Parse()

	// Output the version and exit
	if *flagVersion {
		fmt.Printf("gotrap v%d.%d.%d\n", MajorVersion, MinorVersion, PatchVersion)
		return
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
	stream.Start()
}
