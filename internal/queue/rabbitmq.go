package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

// RabbitMQ constructor
func NewRabbitMQ(cfg *Config) (*RabbitMQ, error) {
	//create tcp connection to rabbitmq
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	//create logical channel inside the tcp connection to rabbitmq
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	//create a queue inside rabbitmq
	_, err = ch.QueueDeclare(
		cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:  conn,
		ch:    ch,
		queue: cfg.QueueName,
	}, nil
}

func (r *RabbitMQ) Channel() *amqp.Channel {
	return r.ch
}

func (r *RabbitMQ) Publish(body []byte) error {
	return r.ch.Publish(
		"",
		r.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQ) Close() error {
	if err := r.ch.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
