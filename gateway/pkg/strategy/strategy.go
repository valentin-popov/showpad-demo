package strategy

// LimitStrategy defines the interface for different rate limiting strategies.
type LimitStrategy interface {
	Accept(userId string, requestsPerSecond float64, path string) bool
}
