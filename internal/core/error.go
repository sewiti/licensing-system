package core

import "errors"

var (
	ErrLicenseExpired        = errors.New("license has expired")
	ErrLicenseSessionExpired = errors.New("license session has expired")

	ErrRateLimitReached = errors.New("rate limit has been reached")
	ErrTimeOutOfSync    = errors.New("time out of sync")
)

// SensitiveError wraps error that shouldn't be exposed to the client directly
// under a generic message.
type SensitiveError struct {
	Message string // Generic message without sensitive information
	err     error  // Underlying sensitive error.
}

// Error returns generic error message without sensitive information.
func (e *SensitiveError) Error() string {
	return e.Message
}

// Unwrap returns underlying error that shouldn't be exposed to the client
// directly.
func (e *SensitiveError) Unwrap() error {
	return e.err
}
