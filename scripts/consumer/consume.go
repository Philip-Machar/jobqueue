package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	//create a tcp connection to rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	//create a channel in the connection
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	defer ch.Close()
	//consume jobs from rabbitmq of the queue stated
	msgs, err := ch.Consume(
		"jobs",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Waiting for jobs...")

	for msg := range msgs {
		log.Println("Received: ", string(msg.Body))

		msg.Ack(false)
	}

}
