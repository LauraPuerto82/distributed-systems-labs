package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitMiddlewareAllowsRequest(t *testing.T) {
	bucket := NewTokenBucket(1, 1)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimitMiddleware(bucket, next)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.Code)
	}

	remaining := response.Header().Get("X-RateLimit-Remaining")

	if remaining != "0" {
		t.Errorf("expected X-RateLimit-Remaining to be 0, got %s", remaining)
	}
}

func TestRateLimitMiddlewareRejectsRequestWhenBucketIsEmpty(t *testing.T) {
	bucket := NewTokenBucket(1, 1)
	bucket.tokens = 0

	nextCalled := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimitMiddleware(bucket, next)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", response.Code)
	}

	if nextCalled {
		t.Error("expected next handler not to be called")
	}

	retryAfter := response.Header().Get("Retry-After")

	if retryAfter != "1" {
		t.Errorf("expected Retry-After to be 1, got %s", retryAfter)
	}

	remaining := response.Header().Get("X-RateLimit-Remaining")

	if remaining != "0" {
		t.Errorf("expected X-RateLimit-Remaining to be 0, got %s", remaining)
	}
}
