package core

import (
	"errors"

	"github.com/sewiti/licensing-system/internal/db"
)

var (
	// License issuer errors
	ErrLicenseIssuerDisabled = errors.New("license issuer is disabled")

	// License errors
	ErrLicenseExpired        = errors.New("license has expired")
	ErrLicenseInactive       = errors.New("license is inactive")
	ErrLicenseSessionExpired = errors.New("license session has expired")

	// Product errors
	ErrProductInactive = errors.New("product is inactive")

	// License session errors
	ErrRateLimitReached = errors.New("rate limit has been reached")
	ErrTimeOutOfSync    = errors.New("time out of sync")

	// Authorization errors
	ErrUserInactive        = errors.New("user is inactive")
	ErrSuperadminImmutable = errors.New("superadmin is immutable")
	ErrInsufficientPerm    = errors.New("insufficient permissions")

	// Database errors
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")

	// Validation errors
	ErrPasswdTooWeak = errors.New("password is too weak")
	ErrInvalidInput  = errors.New("invalid")
	ErrExceedsLimit  = errors.New("exceeds limit")
)

// SensitiveError wraps error that shouldn't be exposed to the client directly
// under a generic message.
type SensitiveError struct {
	Message string // Generic message without sensitive information
	Err     error  // Underlying sensitive error.
}

// Error returns generic error message without sensitive information.
func (e *SensitiveError) Error() string {
	return e.Message
}

// Unwrap returns underlying error that shouldn't be exposed to the client
// directly.
func (e *SensitiveError) Unwrap() error {
	return e.Err
}

// handleErrDB returns database error in a different form.
//  - If error is nil, nil is returned.
//  - If error is db.ErrNotFound, core.ErrNotFound is returned.
//  - If error is db.ErrDuplicate, core.ErrDuplicate is returned.
//  - Other errors are wrapped under core.SensitiveError with a message given.
func handleErrDB(err error, message string) error {
	var sErr *SensitiveError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &sErr):
		sErr.Message = message
		return sErr
	case errors.Is(err, db.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, db.ErrDuplicate):
		return ErrDuplicate
	default:
		return &SensitiveError{
			Message: message,
			Err:     err,
		}
	}
}
