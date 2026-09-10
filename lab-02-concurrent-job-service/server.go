package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type Server struct {
	jobs      chan Job
	nextID    uint64
	processFn func(Job)
}

func NewServer(queueSize int) *Server {
	return &Server{
		jobs: make(chan Job, queueSize),
		processFn: func(job Job) {
			time.Sleep(500 * time.Millisecond)
		},
	}
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Data string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Data == "" {
		http.Error(w, "data is required", http.StatusBadRequest)
		return
	}

	id := atomic.AddUint64(&s.nextID, 1)

	job := Job{
		ID:   strconv.FormatUint(id, 10),
		Data: request.Data,
	}

	s.jobs <- job

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"job_id": job.ID,
	})
}
