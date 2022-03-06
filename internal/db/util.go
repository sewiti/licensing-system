package db

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type selectDecorator func(sq squirrel.SelectBuilder) squirrel.SelectBuilder

func selectPassthrough(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
	return sq
}

func (h *Handler) execInsert(ctx context.Context, sq squirrel.InsertBuilder, scope, action string, id interface{}) error {
	row := sq.QueryRowContext(ctx)
	err := row.Scan(id)
	if err != nil {
		pqErr := &pq.Error{}
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique_violation
				return &Error{err: ErrDuplicate, Scope: scope, Action: action}
			}
		}
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) execDelete(ctx context.Context, sq squirrel.DeleteBuilder, scope, action string) (int, error) {
	res, err := sq.ExecContext(ctx)
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}
	return int(n), nil
}
