package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// 1. Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 2. Open a channel (lightweight connection)
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	// 3. Declare a queue
	q, err := ch.QueueDeclare(
		"jobs", // queue name
		true,   // durable (survives broker restart)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Job message (JSON)
	body := `{"id":"job-1","type":"email","attempt":1}`

	// 5. Publish the job
	err = ch.Publish(
		"", // default exchange
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("job published")
}
