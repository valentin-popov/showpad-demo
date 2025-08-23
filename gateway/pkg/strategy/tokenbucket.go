package strategy

import (
	"sync"
	"time"
)

// TokenBucket strategy uses in-memory token buckets to limit requests.
type TokenBucket struct {
	CurrentTokens map[string]map[string]int // path -> userId -> tokens
	Capacity      int
	LastRefill    map[string]map[string]time.Time // path -> userId -> last refill time
	Created       time.Time
	Mu            sync.Mutex
}

// Accept refills the bucket lazily, based on the elapsed time since the last refill.
// If a token is available, it is consumed and the request is accepted.
func (tb *TokenBucket) Accept(userId string, refillRate float64, path string) bool {
	tb.Mu.Lock()
	defer tb.Mu.Unlock()

	if _, found := tb.LastRefill[path][userId]; !found {
		tb.LastRefill[path][userId] = tb.Created
	}
	now := time.Now()

	elapsedSeconds := now.Sub(tb.LastRefill[path][userId]).Seconds()
	refillTokens := int(elapsedSeconds * refillRate)

	tb.CurrentTokens[path][userId] = min(tb.Capacity, tb.CurrentTokens[path][userId]+refillTokens)
	tb.LastRefill[path][userId] = now

	if tb.CurrentTokens[path][userId] > 0 {
		tb.CurrentTokens[path][userId]--
		return true
	}

	return false
}
