package config

import "errors"

var (
	ErrMissingGateway     = errors.New("gateway config is missing")
	ErrMissingGatewayHost = errors.New("gateway hostname is missing")
	ErrInvalidLogFile     = errors.New("log_file is invalid")
	ErrInvalidDBFile      = errors.New("db_file is invalid")

	ErrInvalidAPIKey      = errors.New("api key is invalid")
	ErrInvalidAPIHostname = errors.New("api hostname is invalid")

	ErrTokenCapacity = errors.New("capacity must be > 0 for route")
	ErrWindowSize    = errors.New("window_size must be > 0 for route")
)
