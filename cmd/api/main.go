package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"jobqueue/internal/jobs"
	"jobqueue/internal/queue"
	"jobqueue/internal/worker"
	"jobqueue/proto/workerpb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func main() {

	// -----------------------------
	// RabbitMQ Publisher Setup
	// -----------------------------
	cfg := queue.Config{
		URL:       "amqp://guest:guest@localhost:5672/",
		QueueName: "jobs",
	}

	rmq, err := queue.NewRabbitMQ(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rmq.Close()

	// -----------------------------
	// Worker Registry (in-memory)
	// -----------------------------
	reg := worker.NewRegistry()

	//registry cleanup
	go func() {
		ticker := time.NewTicker(time.Second * 5)

		for range ticker.C {
			reg.CleanupExpired()
		}
	}()

	// -----------------------------
	// gRPC Server Setup (for workers)
	// -----------------------------
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	workerpb.RegisterWorkerServiceServer(grpcServer, worker.NewGRPCServer(reg))

	go func() {
		log.Println("gRPC server listening on :50051...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// -----------------------------
	// HTTP Server (for job submission)
	// -----------------------------
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Type    string      `json:"type"`
			Payload interface{} `json:"payload"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		job := jobs.Job{
			ID:      uuid.NewString(),
			Type:    req.Type,
			Payload: req.Payload,
		}

		data, err := json.Marshal(job)
		if err != nil {
			http.Error(w, "Failed to encode job: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := rmq.Publish(data); err != nil {
			http.Error(w, "Failed to enqueue job: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	})

	log.Println("Job API listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
