# Lab 01 — Token Bucket Rate Limiter

A concurrent rate limiter implemented in Go using the **token bucket** algorithm.

This lab is part of the Distributed 4 Devs learning path and focuses on concurrency, traffic control, and HTTP rate limiting.

## How it works

The rate limiter uses a token bucket with two configuration values:

- `maxTokens`: maximum number of tokens the bucket can hold.
- `refillRate`: number of tokens generated per second.

Each request consumes one token.

If at least one token is available, the request is allowed. If the bucket does not contain a full token, the request is rejected.

Tokens are refilled lazily based on the elapsed time since the previous refill rather than by running a background process.

For example, with:

```text
maxTokens = 10
refillRate = 2 tokens/second
```

the system allows an initial burst of up to 10 requests and then sustains approximately 2 requests per second as tokens are regenerated.

## Concurrency

`TokenBucket` protects its mutable state with a `sync.Mutex`.

This ensures that concurrent requests cannot simultaneously modify the token count or refill state and produce inconsistent results.

The per-user rate limiter maintains a separate token bucket for each user:

```text
user-a → TokenBucket
user-b → TokenBucket
user-c → TokenBucket
```

The user-to-bucket map has its own mutex because it is shared independently from the state inside each individual bucket.

## HTTP middleware

`RateLimitMiddleware` wraps an `http.Handler` and checks the token bucket before allowing the request to continue.

When a request is allowed, the middleware calls the next handler.

When the limit is exceeded, it returns:

```text
429 Too Many Requests
```

Responses include:

```text
X-RateLimit-Remaining
```

Rejected requests also include:

```text
Retry-After
```

`Retry-After` is calculated from the current fractional token count and refill rate and rounded up to avoid telling clients to retry before a full token is available.

## Per-user rate limiting

`UserRateLimiter` maintains an in-memory mapping:

```go
map[string]*TokenBucket
```

Each user gets an independent bucket, so exhausting one user's rate limit does not affect other users.

Buckets are created lazily when a user is first seen.

## Running the tests

From the lab directory:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

The tests cover:

- Buckets starting at full capacity.
- Token consumption.
- Rejection when the bucket is empty.
- Time-based token refill.
- Maximum bucket capacity.
- `Retry-After` calculation and rounding.
- HTTP requests allowed through the middleware.
- HTTP `429` responses when the limit is exceeded.
- Rate-limit response headers.
- Independent buckets for different users.

## Design decisions

### Buckets start full

New buckets start with `maxTokens` available so the service can accept requests immediately after creation and allow the configured initial burst.

### Lazy refill

Tokens are calculated when `Allow()` is called instead of using a background goroutine.

The number of generated tokens is:

```text
elapsed time × refill rate
```

Fractional tokens are retained internally so sub-second refill rates can be represented accurately.

### Atomic rate-limit result

`Allow()` returns an `AllowResult` containing:

```go
type AllowResult struct {
    Allowed    bool
    Remaining  int
    RetryAfter int
}
```

This keeps the decision and its associated metadata consistent under concurrency. Calling separate methods for `Allow`, `Remaining`, and `RetryAfter` could observe different states if another goroutine modified the bucket between calls.

## Limitations

The per-user bucket registry is stored only in memory.

This means:

- State is lost when the process restarts.
- Multiple application instances would maintain independent limits.
- Inactive user buckets are not currently removed from memory.

A distributed implementation could store shared rate-limit state in a system such as Redis when limits need to be coordinated across multiple application instances.

Periodic cleanup could also be added to remove buckets belonging to inactive users.