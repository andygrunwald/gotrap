package main

import (
	"github.com/streadway/amqp"
	"log"
)

// AmqpInstance reflects a standard instance for an AMQP server
type AmqpInstance struct {
	URI        *amqp.URI
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewAmqpConnection returns a new AMQP connection.
// To establish a connection various credentials like host, port, username, password and vhost are required.
func NewAmqpConnection(host string, port int, username, password, vhost string) *AmqpInstance {
	URI := &amqp.URI{
		Scheme:   "amqp",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Vhost:    vhost,
	}

	instance := &AmqpInstance{
		URI: URI,
	}

	return instance
}

// Connect connects to the AMQP server.
// The credentials are received by the AmqpInstance struct
func (a *AmqpInstance) connect() error {

	var err error

	// Open an AMQP connection
	a.Connection, err = amqp.Dial(a.URI.String())
	if err != nil {
		a.Connection.Close()
		log.Fatalf("> AMQP Connection.open: %s", err)
		return err
	}

	// Open the channel in the new connection
	a.Channel, err = a.Connection.Channel()
	if err != nil {
		a.Channel.Close()
		log.Fatalf("> AMQP Channel.open: %s", err)
		return err
	}

	return nil
}

// DeclareAndBind defines the exchange and queue at the AMQP server.
// We declare our topology on both the publisher and consumer to ensure they
// are the same. This is part of AMQP being a programmable messaging model.
// After declaring we are binding it to be able to receive messages in the queue by the exchange.
func (a AmqpInstance) declareAndBind(exchange, queue, routingKey string) error {

	// Settings:
	//	type: fanout
	// 	durable: false
	//	autoDelete: false
	//	internal: false
	//	noWait: false
	err := a.Channel.ExchangeDeclare(exchange, "fanout", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Exchange.declare: %s", err)
		return err
	}

	// Settings:
	// 	durable: true
	//	autoDelete: false
	//	exclusive: false
	//	noWait: false
	_, err = a.Channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Queue.declare: %v", err)
		return err
	}

	err = a.Channel.QueueBind(queue, routingKey, exchange, false, nil)
	if err != nil {
		log.Fatalf("> AMQP Queue.bind: %v", err)
		return err
	}

	return nil
}
