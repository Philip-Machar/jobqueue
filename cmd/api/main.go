package main

import (
	"encoding/json"
	"jobqueue/internal/jobs"
	"jobqueue/internal/queue"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func main() {
	//initialize publisher config
	cfg := queue.Config{
		URL:       "amqp://guest:guest@localhost:5672/",
		QueueName: "jobs",
	}

	publisher, err := queue.NewPublisher(&cfg)
	if err != nil {
		log.Fatal("", err)
	}
	defer publisher.Close()

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
			http.Error(w, "Invalid request body"+err.Error(), http.StatusBadRequest)
			return
		}

		job := jobs.Job{
			ID:      uuid.NewString(),
			Type:    req.Type,
			Payload: req.Payload,
		}

		data, err := json.Marshal(job)

		if err != nil {
			http.Error(w, "Failed to encode job"+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := publisher.Publish([]byte(data)); err != nil {
			http.Error(w, "Failed to enqueue job"+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(job)
	})

	log.Println("Job API listening at port :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
