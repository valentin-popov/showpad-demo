package limiter

import "fmt"

var (
	errUnauthorized      = fmt.Errorf("unauthorized")
	errNotFound          = fmt.Errorf("not found")
	errRateLimitExceeded = fmt.Errorf("rate limit exceeded")
)
