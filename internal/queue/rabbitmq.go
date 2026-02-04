package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

// RabbitMQ constructor
func NewPublisher(cfg *Config) (*Publisher, error) {
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

	return &Publisher{
		conn:  conn,
		ch:    ch,
		queue: cfg.QueueName,
	}, nil
}

func (r *Publisher) Publish(body []byte) error {
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

func (r *Publisher) Close() error {
	if err := r.ch.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
