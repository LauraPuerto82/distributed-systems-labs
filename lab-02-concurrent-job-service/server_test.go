package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHandleJobsAcceptsValidJob(t *testing.T) {
	server := NewServer(10)

	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{"data":"hello"}`),
	)

	recorder := httptest.NewRecorder()

	server.handleJobs(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusAccepted,
			recorder.Code,
		)
	}

	select {
	case job := <-server.jobs:
		if job.ID != "1" {
			t.Errorf("expected job ID 1, got %s", job.ID)
		}

		if job.Data != "hello" {
			t.Errorf("expected job data hello, got %s", job.Data)
		}

	default:
		t.Fatal("expected job to be added to the queue")
	}
}

func TestHandleJobsRejectsInvalidJSON(t *testing.T) {
	server := NewServer(10)

	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{"data":`),
	)

	recorder := httptest.NewRecorder()

	server.handleJobs(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandleJobsRejectsEmptyData(t *testing.T) {
	server := NewServer(10)

	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{"data":""}`),
	)

	recorder := httptest.NewRecorder()

	server.handleJobs(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandleJobsRejectsNonPostMethod(t *testing.T) {
	server := NewServer(10)

	request := httptest.NewRequest(
		http.MethodGet,
		"/jobs",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.handleJobs(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}
}

func TestHandleJobsBlocksWhenQueueIsFull(t *testing.T) {
	server := NewServer(1)

	// Fill the queue so there is no room for another job.
	server.jobs <- Job{
		ID:   "existing",
		Data: "existing job",
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{"data":"new job"}`),
	)
	rec := httptest.NewRecorder()

	done := make(chan struct{})

	go func() {
		server.handleJobs(rec, req)
		close(done)
	}()

	// The handler should remain blocked because the queue is full.
	select {
	case <-done:
		t.Fatal("expected handler to block while queue is full")

	case <-time.After(50 * time.Millisecond):
		// Expected: the handler is still blocked.
	}

	// Free one slot in the queue.
	<-server.jobs

	// Now the handler should be able to enqueue the new job and finish.
	select {
	case <-done:
		// Expected.

	case <-time.After(1 * time.Second):
		t.Fatal("handler did not finish after queue space became available")
	}

	// Verify that the new job was actually queued.
	select {
	case job := <-server.jobs:
		if job.Data != "new job" {
			t.Errorf("expected job data %q, got %q", "new job", job.Data)
		}

	default:
		t.Fatal("expected new job to be queued")
	}
}

func TestHandleMetricsReturnsCurrentMetrics(t *testing.T) {
	server := NewServer(10)

	server.jobs <- Job{ID: "1", Data: "first"}
	server.jobs <- Job{ID: "2", Data: "second"}

	atomic.StoreUint64(&server.processed, 3)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	server.handleMetrics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var metrics map[string]uint64

	if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if metrics["queued"] != 2 {
		t.Errorf("expected queued 2, got %d", metrics["queued"])
	}

	if metrics["processed"] != 3 {
		t.Errorf("expected processed 3, got %d", metrics["processed"])
	}

	if metrics["rejected"] != 0 {
		t.Errorf("expected rejected 0, got %d", metrics["rejected"])
	}
}
