# distributed-systems-labs

Hands-on labs exploring concurrency, distributed systems, resilience, and system design with Go and Rust.

## Labs

### Lab 01 — Token Bucket Rate Limiter

Concurrent rate limiter implemented in Go using the token bucket algorithm.

Includes time-based token refill, HTTP middleware with rate-limit response headers, and independent in-memory rate limiting per user.

See [`lab-01-token-bucket-rate-limiter`](./lab-01-token-bucket-rate-limiter/) for implementation details, design decisions, tests, and limitations.