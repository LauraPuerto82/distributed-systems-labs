package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerProcessesJob(t *testing.T) {
	server := NewServer(10)

	processed := make(chan Job, 1)

	server.processFn = func(job Job) {
		processed <- job
	}

	server.startWorkers(1)
	defer close(server.jobs)

	expected := Job{
		ID:   "1",
		Data: "hello",
	}

	server.jobs <- expected

	select {
	case job := <-processed:
		if job.ID != expected.ID {
			t.Errorf("expected job ID %s, got %s", expected.ID, job.ID)
		}

		if job.Data != expected.Data {
			t.Errorf("expected job data %s, got %s", expected.Data, job.Data)
		}

	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job to be processed")
	}
}

func TestWorkerPoolProcessesMultipleJobs(t *testing.T) {
	server := NewServer(10)

	processed := make(chan Job, 3)

	server.processFn = func(job Job) {
		processed <- job
	}

	server.startWorkers(2)
	defer close(server.jobs)

	server.jobs <- Job{ID: "1", Data: "first"}
	server.jobs <- Job{ID: "2", Data: "second"}
	server.jobs <- Job{ID: "3", Data: "third"}

	received := make(map[string]bool)

	for i := 0; i < 3; i++ {
		select {
		case job := <-processed:
			received[job.ID] = true

		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for jobs to be processed")
		}
	}

	if !received["1"] {
		t.Error("expected job 1 to be processed")
	}

	if !received["2"] {
		t.Error("expected job 2 to be processed")
	}

	if !received["3"] {
		t.Error("expected job 3 to be processed")
	}
}

func TestWorkersFinishAfterJobsChannelIsClosed(t *testing.T) {
	server := NewServer(10)

	processed := make(chan Job, 2)

	server.processFn = func(job Job) {
		processed <- job
	}

	server.startWorkers(2)

	server.jobs <- Job{ID: "1", Data: "first"}
	server.jobs <- Job{ID: "2", Data: "second"}

	close(server.jobs)

	done := make(chan struct{})

	go func() {
		server.workers.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected: all workers finished.

	case <-time.After(1 * time.Second):
		t.Fatal("workers did not finish after jobs channel was closed")
	}

	if len(processed) != 2 {
		t.Errorf("expected 2 jobs to be processed, got %d", len(processed))
	}
}

func TestWorkerIncrementsProcessedCount(t *testing.T) {
	server := NewServer(10)

	server.processFn = func(job Job) {}

	server.startWorkers(1)

	server.jobs <- Job{
		ID:   "1",
		Data: "test",
	}

	close(server.jobs)

	server.workers.Wait()

	got := atomic.LoadUint64(&server.processed)

	if got != 1 {
		t.Errorf("expected processed count 1, got %d", got)
	}
}
