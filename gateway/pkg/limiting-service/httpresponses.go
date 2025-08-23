package limiter

const (
	respUnauthorized      = "{error: 'unauthorized'}"
	respBadRequest        = "{error: 'bad request'}"
	respNotFound          = "{error: 'not found'}"
	respRateLimitExceeded = "{error: 'rate limit exceeded'}"
	respInternalServer    = "{error: 'internal server error'}"
	respSuccess           = "{success: true}"
)
