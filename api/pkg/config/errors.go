package config

import "errors"

var (
	ErrInvalidAPIKey      = errors.New("api key is invalid")
	ErrInvalidAPIHostname = errors.New("api hostname is invalid")
)
