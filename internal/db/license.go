package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const licenseTable = "license"

func (h *Handler) InsertLicense(ctx context.Context, l *model.License) error {
	const action = "Insert"
	sq := h.sq.Insert(licenseTable).
		SetMap(map[string]interface{}{
			"id":           l.ID[:],
			"key":          l.Key[:],
			"note":         l.Note,
			"data":         l.Data,
			"max_sessions": l.MaxSessions,
			"valid_until":  l.ValidUntil,
			"created":      l.Created,
			"updated":      l.Updated,
			"issuer_id":    l.IssuerID,
		})

	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: licenseTable, Action: action}
	}
	return nil
}

func (h *Handler) SelectAllLicensesByIssuerID(ctx context.Context, licenseIssuerID int) ([]*model.License, error) {
	return h.selectLicenses(ctx, "SelectAllByIssuerID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"issuer_id": licenseIssuerID,
			})
		})
}

func (h *Handler) SelectLicenseByID(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	return h.selectLicense(ctx, "SelectByID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"id": licenseID[:],
			})
		})
}

func (h *Handler) selectLicense(ctx context.Context, action string, d selectDecorator) (*model.License, error) {
	ll, err := h.selectLicenses(ctx, action, d)
	if err != nil {
		return nil, err
	}
	if len(ll) == 0 {
		return nil, &Error{err: ErrNotFound, Scope: licenseTable, Action: action}
	}
	return ll[0], nil
}

func (h *Handler) selectLicenses(ctx context.Context, action string, d selectDecorator) ([]*model.License, error) {
	const scope = licenseTable

	sq := h.sq.Select(
		"id",
		"key",
		"note",
		"data",
		"max_sessions",
		"valid_until",
		"created",
		"updated",
		"issuer_id",
	).From(scope)

	rows, err := d(sq).QueryContext(ctx)
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	defer rows.Close()

	var ll []*model.License
	for rows.Next() {
		var id []byte
		var key []byte
		l := &model.License{}
		err = rows.Scan(
			&id,
			&key,
			&l.Note,
			&l.Data,
			&l.MaxSessions,
			&l.ValidUntil,
			&l.Created,
			&l.Updated,
			&l.IssuerID,
		)
		if err != nil {
			return nil, &Error{err: err, Scope: scope, Action: action}
		}
		l.ID = (*[32]byte)(id)
		l.Key = (*[32]byte)(key)
		ll = append(ll, l)
	}

	err = rows.Err()
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	return ll, nil
}

func (h *Handler) SelectLicensesCountByIssuerID(ctx context.Context, licenseIssuerID int) (int, error) {
	const (
		scope  = licenseTable
		action = "SelectCountByIssuerID"
	)
	sq := h.sq.Select("COUNT(*)").
		From(scope).
		Where(squirrel.Eq{
			"issuer_id": licenseIssuerID,
		})

	row := sq.QueryRowContext(ctx)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}
	return count, nil
}

func (h *Handler) UpdateLicense(ctx context.Context, licenseID *[32]byte, update map[string]interface{}) error {
	const (
		action = "Update"
		scope  = licenseTable
	)
	sq := h.sq.Update(scope).
		SetMap(update).
		Where(squirrel.Eq{
			"id": licenseID[:],
		})

	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) DeleteLicenseByID(ctx context.Context, licenseID *[32]byte) (int, error) {
	const scope = licenseTable
	sq := h.sq.Delete(scope).
		Where(squirrel.Eq{
			"id": licenseID[:],
		})
	return h.execDelete(ctx, sq, scope, "DeleteByID")
}
