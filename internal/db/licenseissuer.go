package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const licenseIssuerTable = "license_issuer"

func (h *Handler) InsertLicenseIssuer(ctx context.Context, li *model.LicenseIssuer) (int, error) {
	const (
		action = "Insert"
		scope  = licenseIssuerTable
	)
	sq := h.sq.Insert(scope).
		SetMap(map[string]interface{}{
			"active":        li.Active,
			"username":      li.Username,
			"password_hash": li.PasswordHash,
			"max_licenses":  li.MaxLicenses,
			"created":       li.Created,
			"updated":       li.Updated,
		}).Suffix("RETURNING id")

	var id int
	return id, h.execInsert(ctx, sq, scope, action, &id)
}

func (h *Handler) SelectAllLicenseIssuers(ctx context.Context) ([]*model.LicenseIssuer, error) {
	return h.selectLicenseIssuers(ctx, "SelectAll", selectPassthrough)
}

func (h *Handler) SelectLicenseIssuerByUsername(ctx context.Context, licenseIssuerUsername string) (*model.LicenseIssuer, error) {
	return h.selectLicenseIssuer(ctx, "SelectByUsername",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"username": licenseIssuerUsername,
			})
		})
}

func (h *Handler) SelectLicenseIssuerByID(ctx context.Context, licenseIssuerID int) (*model.LicenseIssuer, error) {
	return h.selectLicenseIssuer(ctx, "SelectByID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"id": licenseIssuerID,
			})
		})
}

func (h *Handler) selectLicenseIssuer(ctx context.Context, action string, d selectDecorator) (*model.LicenseIssuer, error) {
	lii, err := h.selectLicenseIssuers(ctx, action, d)
	if err != nil {
		return nil, err
	}
	if len(lii) == 0 {
		return nil, &Error{err: ErrNotFound, Scope: licenseIssuerTable, Action: action}
	}
	return lii[0], nil
}

func (h *Handler) selectLicenseIssuers(ctx context.Context, action string, d selectDecorator) ([]*model.LicenseIssuer, error) {
	const scope = licenseIssuerTable

	sq := h.sq.Select(
		"id",
		"active",
		"username",
		"password_hash",
		"max_licenses",
		"created",
		"updated",
	).From(scope)

	rows, err := d(sq).QueryContext(ctx)
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	defer rows.Close()

	var lii []*model.LicenseIssuer
	for rows.Next() {
		li := &model.LicenseIssuer{}
		err = rows.Scan(
			&li.ID,
			&li.Active,
			&li.Username,
			&li.PasswordHash,
			&li.MaxLicenses,
			&li.Created,
			&li.Updated,
		)
		if err != nil {
			return nil, &Error{err: err, Scope: scope, Action: action}
		}
		lii = append(lii, li)
	}

	err = rows.Err()
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	return lii, nil
}

func (h *Handler) UpdateLicenseIssuer(ctx context.Context, licenseIssuerID int, update map[string]interface{}) error {
	const (
		action = "Update"
		scope  = licenseIssuerTable
	)
	sq := h.sq.Update(scope).
		SetMap(update).
		Where(squirrel.Eq{
			"id": licenseIssuerID,
		})

	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) UpdateLicenseIssuerByUsername(ctx context.Context, username string, update map[string]interface{}) error {
	const (
		action = "UpdateByUsername"
		scope  = licenseIssuerTable
	)
	sq := h.sq.Update(scope).
		SetMap(update).
		Where(squirrel.Eq{
			"username": username,
		})

	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) DeleteLicenseIssuerByID(ctx context.Context, licenseIssuerID int) (int, error) {
	const scope = licenseIssuerTable
	sq := h.sq.Delete(scope).
		Where(squirrel.Eq{
			"id": licenseIssuerID,
		})
	return h.execDelete(ctx, sq, scope, "DeleteByID")
}
