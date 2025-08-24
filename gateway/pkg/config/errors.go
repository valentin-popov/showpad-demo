package config

import "errors"

var (
	ErrMissingGateway        = errors.New("gateway config is missing")
	ErrMissingGatewayAddress = errors.New("gateway address is missing")
	ErrInvalidLogFile        = errors.New("log_file is invalid")
	ErrInvalidDBFile         = errors.New("db_file is invalid")

	ErrInvalidAPIKey     = errors.New("api key is invalid")
	ErrInvalidAPIAddress = errors.New("api address is invalid")

	ErrTokenCapacity = errors.New("capacity must be > 0 for route")
	ErrWindowSize    = errors.New("window_size must be > 0 for route")
)
