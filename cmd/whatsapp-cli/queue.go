package main

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

// QueueManager handles RabbitMQ operations
type QueueManager struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  QueueConfig
}

type QueueConfig struct {
	URL      string `mapstructure:"url"`
	Exchange string `mapstructure:"exchange"`
	Queues   struct {
		Incoming      string `mapstructure:"incoming"`
		Outgoing      string `mapstructure:"outgoing"`
		Media         string `mapstructure:"media"`
		AI            string `mapstructure:"ai"`
		Notifications string `mapstructure:"notifications"`
		System        string `mapstructure:"system"`
	} `mapstructure:"queues"`
}

func NewQueueManager(config QueueConfig) (*QueueManager, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		config.Exchange, // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Declare queues
	queues := []string{
		config.Queues.Incoming,
		config.Queues.Outgoing,
		config.Queues.Media,
		config.Queues.AI,
		config.Queues.Notifications,
		config.Queues.System,
	}

	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}

		// Bind queue to exchange
		err = ch.QueueBind(
			queue,           // queue name
			queue+".*",      // routing key
			config.Exchange, // exchange
			false,
			nil,
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, err
		}
	}

	return &QueueManager{
		conn:    conn,
		channel: ch,
		config:  config,
	}, nil
}

func (qm *QueueManager) PublishMessage(routingKey, message string) error {
	return qm.channel.Publish(
		qm.config.Exchange, // exchange
		routingKey,         // routing key
		false,              // mandatory
		false,              // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			Timestamp:   time.Now(),
		})
}

func (qm *QueueManager) ConsumeMessages(queueName string, handler func(amqp091.Delivery)) error {
	msgs, err := qm.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			handler(d)
			d.Ack(false) // acknowledge message
		}
	}()

	return nil
}

func (qm *QueueManager) Close() {
	if qm.channel != nil {
		qm.channel.Close()
	}
	if qm.conn != nil {
		qm.conn.Close()
	}
}

// Global queue manager
var queueMgr *QueueManager
