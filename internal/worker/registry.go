package worker

import (
	"log"
	"sync"
	"time"
)

const workerTTL = time.Second * 15

type Registry struct {
	mu      sync.Mutex
	workers map[string]time.Time
}

func NewRegistry() *Registry {
	return &Registry{
		workers: make(map[string]time.Time),
	}
}

func (r *Registry) Register(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[id] = time.Now()
}

func (r *Registry) Heartbeat(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[id] = time.Now()
}

func (r *Registry) List() map[string]time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := make(map[string]time.Time)
	for k, v := range r.workers {
		copy[k] = v
	}
	return copy
}

func (r *Registry) CleanupExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, lastSeen := range r.workers {
		if now.Sub(lastSeen) > workerTTL {
			log.Printf("Worker %s expired — removing", id)
			delete(r.workers, id)
		}
	}
}
