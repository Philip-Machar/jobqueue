package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"jobqueue/internal/jobs"
	"jobqueue/internal/queue"
	"jobqueue/proto/workerpb"
)

const maxAttempts = 3

func main() {

	// -----------------------------------
	// Connect to RabbitMQ
	// -----------------------------------
	cfg := queue.Config{
		URL:       "amqp://guest:guest@localhost:5672/",
		QueueName: "jobs",
	}

	rmq, err := queue.NewRabbitMQ(&cfg)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer rmq.Close()

	// -----------------------------------
	// Connect to API via gRPC
	// -----------------------------------
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to API gRPC server: %v", err)
	}
	defer conn.Close()

	client := workerpb.NewWorkerServiceClient(conn)

	workerID := uuid.NewString()

	// Register worker
	_, err = client.Register(context.Background(), &workerpb.RegisterRequest{WorkerId: workerID})
	if err != nil {
		log.Fatalf("failed to register worker: %v", err)
	}

	log.Printf("Worker registered with ID: %s", workerID)

	// -----------------------------------
	// Start Heartbeat Loop
	// -----------------------------------
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			_, err := client.Heartbeat(context.Background(), &workerpb.HeartbeatRequest{WorkerId: workerID})
			if err != nil {
				log.Println("heartbeat failed:", err)
				continue
			}
			log.Println("heartbeat sent")
		}
	}()

	// -----------------------------------
	// RabbitMQ Consumer Setup
	// -----------------------------------
	if err := rmq.Channel().Qos(1, 0, false); err != nil {
		log.Fatalf("Failed to set Qos %v", err)
	}

	msgs, err := rmq.Channel().Consume(
		cfg.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to consume messages: %v", err)
	}

	log.Println("Worker started and consuming jobs")
	log.Println("Waiting for job messages...")

	for msg := range msgs {
		log.Println("Received message from queue")
		handleMessage(msg, rmq)
	}
}

func handleMessage(msg amqp.Delivery, rmq *queue.RabbitMQ) {
	var job jobs.Job

	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Println("invalid job payload, discarded")
		_ = msg.Nack(false, false)
		return
	}

	log.Printf(
		"processing job %s (type=%s, attempt=%d)",
		job.ID,
		job.Type,
		job.Attempts,
	)

	// Simulated failure logic
	if job.Attempts < maxAttempts-1 {
		job.Attempts++

		log.Printf(
			"job %s failed, retrying (%d/%d)",
			job.ID,
			job.Attempts,
			maxAttempts,
		)

		data, err := json.Marshal(job)
		if err != nil {
			log.Println("failed to re-marshal job, discarding")
			_ = msg.Nack(false, false)
			return
		}

		// discard original message
		_ = msg.Nack(false, false)

		// re-enqueue
		if err := rmq.Publish(data); err != nil {
			log.Println("failed to re-publish job:", err)
		}
		return
	}

	log.Printf("job %s succeeded", job.ID)
	_ = msg.Ack(false)
}
