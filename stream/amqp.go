package stream

import (
	"encoding/json"
	"github.com/andygrunwald/gotrap/config"
	"github.com/andygrunwald/gotrap/gerrit"
	"github.com/andygrunwald/gotrap/github"
	"github.com/streadway/amqp"
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
	var wg sync.WaitGroup
	// Limit number of concurrent patch requests here with a semaphore
	sem := make(chan bool, s.Config.Gotrap.Concurrent)

	// We have to do this in a loop, to reconnect to rabbitmq automatically
	// This connection times out sometimes.
	for {
		// If we don`t get a AMQP connection we can exit here
		// Without AMQP connection gotrap is useless
		if err := s.Connect(); err != nil {
			return err
		}
		defer s.Connection.Close()

		// Declare AMQP exchange and queue and bind them together :)
		// If this will fail we can exit here with the same reason like above
		// Without queue gotrap is useless
		if err := s.DeclareAndBind(&s.Config.Amqp); err != nil {
			return err
		}

		// Get the consumer channel to get all messages
		messages, err := s.Channel.Consume(s.Config.Amqp.Queue, s.Config.Amqp.Identifier, false, false, false, false, nil)
		if err != nil {
			return err
		}

		// Get new messages by the AMQP broker
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

				// Acknowledge message if we get this
				// We do this, because at the end of proceeding we might lost the connection to AMQP server
				// I know this is wrong, but currently the reconnection does not work correctly :(
				// TODO: Fix this later and Ack message if its done
				event.Ack(false)

				// Convert the AMQP into a Gerrit message
				var change gerrit.Message
				err := json.Unmarshal(event.Body, &change)
				// If we can`t read the message, we will exit here
				if err != nil {
					return
				}

				// Build the main data structure and start working on the message :)
				gotrap := Gotrap{
					githubClient: *github.NewGithubClient(&s.Config.Github),
					gerritClient: *gerrit.NewGerritClient(&s.Config.Gerrit),
					config:       s.Config,
					Message:      change,
				}
				gotrap.TakeAction()
			}()
		}
	}

	wg.Wait()
	return nil
}

// Connect connects to the AMQP server.
// The credentials are received by the AmqpInstance struct
func (s *AmqpStream) Connect() error {

	s.URI = &amqp.URI{
		Scheme:   "amqp",
		Host:     s.Config.Amqp.Host,
		Port:     s.Config.Amqp.Port,
		Username: s.Config.Amqp.Username,
		Password: s.Config.Amqp.Password,
		Vhost:    s.Config.Amqp.VHost,
	}

	var err error

	// Open an AMQP connection
	s.Connection, err = amqp.Dial(s.URI.String())
	if err != nil {
		return err
	}

	// Open the channel in the new connection
	s.Channel, err = s.Connection.Channel()
	if err != nil {
		s.Connection.Close()
		return err
	}

	return nil
}

// DeclareAndBind defines the exchange and queue at the AMQP server.
// We declare our topology on both the publisher and consumer to ensure they
// are the same. This is part of AMQP being a programmable messaging model.
// After declaring we are binding it to be able to receive messages in the queue by the exchange.
func (s *AmqpStream) DeclareAndBind(config *config.AmqpConfiguration) error {

	// Settings:
	//	type: fanout
	// 	durable: false
	//	autoDelete: false
	//	internal: false
	//	noWait: false
	err := s.Channel.ExchangeDeclare(config.Exchange, "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// Settings:
	// 	durable: true
	//	autoDelete: false
	//	exclusive: false
	//	noWait: false
	_, err = s.Channel.QueueDeclare(config.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = s.Channel.QueueBind(config.Queue, config.RoutingKey, config.Exchange, false, nil)
	if err != nil {
		return err
	}

	return nil
}
