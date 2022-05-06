package core

import (
	"errors"
	"testing"

	"github.com/sewiti/licensing-system/internal/db"
	"github.com/stretchr/testify/assert"
)

func Test_handleErrDB(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
		want    error
	}{
		{
			name:    "nil",
			message: "message",
			err:     nil,
			want:    nil,
		},
		{
			name:    "duplicate",
			message: "issuer duplicate name",
			err:     db.ErrDuplicate,
			want:    ErrDuplicate,
		},
		{
			name:    "not found",
			message: "issuer not found",
			err:     db.ErrNotFound,
			want:    ErrNotFound,
		},
		{
			name:    "change message",
			message: "new message",
			err:     &SensitiveError{Err: ErrExceedsLimit, Message: "old message"},
			want:    &SensitiveError{Err: ErrExceedsLimit, Message: "new message"},
		},
		{
			name:    "other",
			message: "other message",
			err:     errors.New("other error"),
			want:    &SensitiveError{Err: errors.New("other error"), Message: "other message"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, handleErrDB(tt.err, tt.message))
		})
	}
}
