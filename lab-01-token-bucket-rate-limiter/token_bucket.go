package main

import (
	"math"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

type AllowResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter int
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) retryAfter() int {
	if tb.refillRate <= 0 {
		return 0
	}

	secondsUntilNextToken := (1 - tb.tokens) / tb.refillRate

	return int(math.Ceil(secondsUntilNextToken))
}

func (tb *TokenBucket) Allow() AllowResult {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens -= 1
		return AllowResult{
			Allowed:    true,
			Remaining:  int(tb.tokens),
			RetryAfter: 0,
		}
	}

	return AllowResult{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: tb.retryAfter(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefill = now
}
