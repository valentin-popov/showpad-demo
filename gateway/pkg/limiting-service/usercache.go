package limiter

import "time"

// UserCache is a simple in-memory cache for user request rates with TTL.
type UserCache struct {
	data map[string]userData
	ttl  time.Duration
}

type userData struct {
	userId    string
	reqPerSec float64
	created   time.Time
}

// Add adds a user with their request rate to the cache.
// If the user already exists, their data is updated.
func (cache *UserCache) Add(userId string, reqPerSec float64) {
	cache.data[userId] = userData{
		userId:    userId,
		reqPerSec: reqPerSec,
		created:   time.Now(),
	}
}

// GetRate returns the user request rate, if found and not expired
// Otherwise, 0 is returned
func (cache *UserCache) GetRate(userId string) float64 {
	data, exists := cache.data[userId]

	if !exists {
		return 0
	}

	if time.Since(data.created) > cache.ttl {
		delete(cache.data, userId)
		return 0
	}
	return data.reqPerSec
}
