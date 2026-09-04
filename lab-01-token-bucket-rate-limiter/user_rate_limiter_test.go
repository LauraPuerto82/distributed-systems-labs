package main

import "testing"

func TestGetBucketReturnsSameBucketForSameUser(t *testing.T) {
	limiter := NewUserRateLimiter(10, 2)

	first := limiter.getBucket("laura")
	second := limiter.getBucket("laura")

	if first != second {
		t.Error("expected same bucket for the same user")
	}
}

func TestGetBucketReturnsDifferentBucketsForDifferentUsers(t *testing.T) {
	limiter := NewUserRateLimiter(10, 2)

	lauraBucket := limiter.getBucket("laura")
	pedroBucket := limiter.getBucket("pedro")

	if lauraBucket == pedroBucket {
		t.Error("expected different buckets for different users")
	}
}

func TestUserRateLimiterLimitsUsersIndependently(t *testing.T) {
	limiter := NewUserRateLimiter(1, 0)

	firstLaura := limiter.Allow("laura")
	secondLaura := limiter.Allow("laura")
	firstPedro := limiter.Allow("pedro")

	if !firstLaura.Allowed {
		t.Error("expected Laura's first request to be allowed")
	}

	if secondLaura.Allowed {
		t.Error("expected Laura's second request to be rejected")
	}

	if !firstPedro.Allowed {
		t.Error("expected Pedro's first request to be allowed")
	}
}
