package config

import "errors"

var (
	ErrInvalidAPIKey     = errors.New("api key is invalid")
	ErrInvalidAPIAddress = errors.New("api address is invalid")
)
