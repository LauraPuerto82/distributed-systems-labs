package main

import "sync"

type UserRateLimiter struct {
	buckets    map[string]*TokenBucket
	maxTokens  float64
	refillRate float64
	mu         sync.Mutex
}

func NewUserRateLimiter(maxTokens, refillRate float64) *UserRateLimiter {
	return &UserRateLimiter{
		buckets:    make(map[string]*TokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (rl *UserRateLimiter) getBucket(userID string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[userID]

	if exists {
		return bucket
	}

	bucket = NewTokenBucket(rl.maxTokens, rl.refillRate)
	rl.buckets[userID] = bucket

	return bucket
}

func (rl *UserRateLimiter) Allow(userID string) AllowResult {
	bucket := rl.getBucket(userID)
	return bucket.Allow()
}
