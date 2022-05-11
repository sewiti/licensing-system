package db

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate")
)

type Error struct {
	Scope  string
	Action string
	err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("db: %s: %s: %v", e.Scope, e.Action, e.err)
}

func (e *Error) Unwrap() error {
	return e.err
}
