package strategy

// FixedWindow strategy uses an SQL table to track request counts per user and path.
type FixedWindow struct {
}

func (fw *FixedWindow) Accept(userId string, requestsPerSecond float64, path string) bool {

	return true
}
