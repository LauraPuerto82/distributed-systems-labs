package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
