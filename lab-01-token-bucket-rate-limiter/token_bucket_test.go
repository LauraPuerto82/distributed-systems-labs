package main

import (
	"testing"
	"time"
)

func TestNewTokenBucketStartsFull(t *testing.T) {
	bucket := NewTokenBucket(10, 2)

	if bucket.tokens != 10 {
		t.Errorf("expected bucket to start with 10 tokens, got %f", bucket.tokens)
	}
}

func TestAllowConsumesOneToken(t *testing.T) {
	bucket := NewTokenBucket(10, 2)

	allowed := bucket.Allow()

	if !allowed {
		t.Error("expected request to be allowed")
	}

	if bucket.tokens != 9 {
		t.Errorf("expected 9 tokens remaining, got %f", bucket.tokens)
	}
}

func TestAllowRejectsWhenBucketIsEmpty(t *testing.T) {
	bucket := NewTokenBucket(2, 1)

	if !bucket.Allow() {
		t.Error("expected first request to be allowed")
	}

	if !bucket.Allow() {
		t.Error("expected second request to be allowed")
	}

	if bucket.Allow() {
		t.Error("expected third request to be rejected")
	}
}

func TestRefillAddsTokensBasedOnElapsedTime(t *testing.T) {
	bucket := NewTokenBucket(10, 2)
	bucket.tokens = 0
	bucket.lastRefill = time.Now().Add(-2 * time.Second)

	bucket.refill()

	if bucket.tokens < 4 || bucket.tokens >= 4.1 {
		t.Errorf("expected approximately 4 tokens, got %f", bucket.tokens)
	}
}

func TestRefillDoesNotExceedMaxTokens(t *testing.T) {
	bucket := NewTokenBucket(10, 2)
	bucket.tokens = 8
	bucket.lastRefill = time.Now().Add(-10 * time.Second)

	bucket.refill()

	if bucket.tokens != 10 {
		t.Errorf("expected bucket to be capped at 10 tokens, got %f", bucket.tokens)
	}
}
