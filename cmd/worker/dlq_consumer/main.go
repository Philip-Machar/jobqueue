package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DeadJob struct {
	ID       string `json:"id"`
	Attempts int    `json:"attempts"`
	Payload  any    `json:"payload"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("DLQ connect failed:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("DLQ channel failed:", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"jobs.dlq",
		"dlq-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("DLQ consume failed:", err)
	}

	log.Println("DLQ consumer running")

	for msg := range msgs {
		var job DeadJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			log.Println("Invalid DLQ message:", err)
			msg.Ack(false)
			continue
		}

		log.Printf(
			"DEAD JOB: id=%s attempts=%d payload=%v",
			job.ID,
			job.Attempts,
			job.Payload,
		)

		// Ack so it doesn't loop forever
		msg.Ack(false)
	}
}
