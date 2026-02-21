package worker

import (
	"log"
	"sync"
	"time"
)

const workerTTL = time.Second * 15

type workerInfo struct {
	LastSeen time.Time
	Load     int32
}

type Registry struct {
	mu      sync.Mutex
	workers map[string]*workerInfo
}

func NewRegistry() *Registry {
	return &Registry{
		workers: make(map[string]*workerInfo),
	}
}

func (r *Registry) Register(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[id] = &workerInfo{LastSeen: time.Now(), Load: 0}
}

func (r *Registry) Heartbeat(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, ok := r.workers[id]; ok {
		w.LastSeen = time.Now()
	}
}

func (r *Registry) UpdateLoad(id string, load int32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, ok := r.workers[id]; ok {
		w.Load = load
		w.LastSeen = time.Now()
	}
}

func (r *Registry) List() map[string]workerInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	copy := make(map[string]workerInfo)

	for id, info := range r.workers {
		copy[id] = *info
	}
	return copy
}

func (r *Registry) CleanupExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, info := range r.workers {
		if now.Sub(info.LastSeen) > workerTTL {
			log.Printf("Worker %s expired — removing", id)
			delete(r.workers, id)
		}
	}
}
