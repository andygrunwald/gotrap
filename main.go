package main

import (
	"flag"
	"fmt"
	"log"
	//"runtime"
	"sync"
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

	log.Println("Hey, nice to meet you. Just wait a second. I will start up.")
	defer log.Println("Our job is done. We have to go.")

	// Bootstrap configuration file
	var config Configuration
	config.init(flagConfigFile)

	// We have to do this in a loop, to reconnect to rabbitmq automatically
	// This connection times out sometimes.
	for {

		// Build the AMQP connection
		amqp := NewAmqpConnection(config.Amqp.Host, config.Amqp.Port, config.Amqp.Username, config.Amqp.Password, config.Amqp.VHost)

		// If we don`t get a AMQP connection we can exit here
		// Without AMQP connection gotrap is useless
		if err := amqp.connect(); err != nil {
			log.Fatalf("> AMQP connection not available: %v", err)
		}
		defer amqp.Connection.Close()

		// Declare AMQP exchange and queue and bind them together :)
		// If this will fail we can exit here with the same reason like above
		// Without queue gotrap is useless
		if err := amqp.declareAndBind(config.Amqp.Exchange, config.Amqp.Queue, config.Amqp.RoutingKey); err != nil {
			log.Fatalf("> AMQP Declare and bind: %v", err)
		}

		// Get the consumer channel to get all messages
		messages, err := amqp.Channel.Consume(config.Amqp.Queue, config.Amqp.Identifier, false, false, false, false, nil)
		if err != nil {
			log.Fatalf("> AMQP Basic.consume: %v", err)
		}

		// Bootstrap a waitgroup
		// With this we are running as long as the go routines run
		var wg sync.WaitGroup
		wg.Add(1)

		// Limit number of concurrent patch requests here with a semaphore
		sem := make(chan bool, config.Gotrap.Concurrent)

		// Start main go routine to receive messages by the AMQP broker
		go func() {
			defer wg.Done()

			// Get new messages
			for event := range messages {
				// Semaphore! Fill it
				sem <- true
				wg.Add(1)

				// One go routine per message
				go func() {
					defer func() {
						// Semaphore! Release it if this message was handled
						<-sem
						wg.Done()
					}()

					// Bootstrap the Github and Gerrit client ...
					githubClient := *NewGithubClient(&config.Github)
					gerritClient := *NewGerritInstance(&config.Gerrit)

					// ... and start handle the message!
					handleNewMessage(githubClient, gerritClient, config, event)
				}()
			}
		}()

		wg.Wait()
	}

}
