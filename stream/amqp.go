package stream

import (
	"fmt"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/andygrunwald/gotrap/github"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

type AmqpStream struct {
	URI        *amqp.URI
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Config     *config.Configuration
}

func init() {
	Streams[StreamAmqp] = new(AmqpStream)
}

func (s *AmqpStream) Initialize(config *config.Configuration) {
	s.Config = config
}

func (s *AmqpStream) Start() error {
	// We have to do this in a loop, to reconnect to rabbitmq automatically
	// This connection times out sometimes.
	for {

		// Build the AMQP connection
		s.URI = newAmqpConnection(&s.Config.Amqp)

		// If we don`t get a AMQP connection we can exit here
		// Without AMQP connection gotrap is useless
		if err := s.connect(); err != nil {
			log.Fatalf("> AMQP connection not available: %v", err)
		}
		defer s.Connection.Close()

		// Declare AMQP exchange and queue and bind them together :)
		// If this will fail we can exit here with the same reason like above
		// Without queue gotrap is useless
		if err := s.declareAndBind(&s.Config.Amqp); err != nil {
			log.Fatalf("> AMQP Declare and bind: %v", err)
		}

		// Get the consumer channel to get all messages
		messages, err := s.Channel.Consume(s.Config.Amqp.Queue, s.Config.Amqp.Identifier, false, false, false, false, nil)
		if err != nil {
			log.Fatalf("> AMQP Basic.consume: %v", err)
		}

		// Bootstrap a waitgroup
		// With this we are running as long as the go routines run
		var wg sync.WaitGroup
		wg.Add(1)

		// Limit number of concurrent patch requests here with a semaphore
		sem := make(chan bool, s.Config.Gotrap.Concurrent)

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
					githubClient := *github.NewGithubClient(&s.Config.Github)
					gerritClient := *gerrit.NewGerritClient(&s.Config.Gerrit)

					// ... and start handle the message!
					handleNewMessage(githubClient, gerritClient, s.Config, event)
				}()
			}
		}()

		wg.Wait()
	}
}

// NewAmqpConnection returns a new AMQP connection.
// To establish a connection various credentials like host, port, username, password and vhost are required.
func newAmqpConnection(config *config.AmqpConfiguration) *amqp.URI {
	uri := &amqp.URI{
		Scheme:   "amqp",
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Vhost:    config.VHost,
	}

	return uri
}

// Connect connects to the AMQP server.
// The credentials are received by the AmqpInstance struct
func (s *AmqpStream) connect() error {

	var err error

	// Open an AMQP connection
	s.Connection, err = amqp.Dial(s.URI.String())
	if err != nil {
		s.Connection.Close()
		log.Fatalf("> AMQP Connection.open: %s", err)
		return err
	}

	// Open the channel in the new connection
	s.Channel, err = s.Connection.Channel()
	if err != nil {
		s.Channel.Close()
		log.Fatalf("> AMQP Channel.open: %s", err)
		return err
	}

	return nil
}

// DeclareAndBind defines the exchange and queue at the AMQP server.
// We declare our topology on both the publisher and consumer to ensure they
// are the same. This is part of AMQP being a programmable messaging model.
// After declaring we are binding it to be able to receive messages in the queue by the exchange.
func (s *AmqpStream) declareAndBind(config *config.AmqpConfiguration) error {

	// Settings:
	//	type: fanout
	// 	durable: false
	//	autoDelete: false
	//	internal: false
	//	noWait: false
	err := s.Channel.ExchangeDeclare(config.Exchange, "fanout", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Exchange.declare: %s", err)
		return err
	}

	// Settings:
	// 	durable: true
	//	autoDelete: false
	//	exclusive: false
	//	noWait: false
	_, err = s.Channel.QueueDeclare(config.Queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Queue.declare: %v", err)
		return err
	}

	err = s.Channel.QueueBind(config.Queue, config.RoutingKey, config.Exchange, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Queue.bind: %v", err)
		return err
	}

	return nil
}
