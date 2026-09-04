package main

import (
	"net/http"
	"strconv"
)

func RateLimitMiddleware(bucket *TokenBucket, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := bucket.Allow()

		w.Header().Set(
			"X-RateLimit-Remaining",
			strconv.Itoa(result.Remaining),
		)

		if !result.Allowed {
			w.Header().Set(
				"Retry-After",
				strconv.Itoa(result.RetryAfter),
			)

			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
