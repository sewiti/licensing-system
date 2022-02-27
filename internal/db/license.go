package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const licensesTable = "licenses"

func (h *Handler) InsertLicense(ctx context.Context, l *model.License) error {
	const action = "Insert"
	sq := h.sq.Insert(licensesTable).
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
		return &Error{err: err, Scope: licensesTable, Action: action}
	}
	return nil
}

func (h *Handler) SelectLicenseByID(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	const action = "SelectByID"
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
	).
		From(licensesTable).
		Where(squirrel.Eq{
			"id": licenseID[:],
		})

	var id []byte
	var key []byte
	l := &model.License{}
	err := h.scanRow(sq.QueryRowContext(ctx),
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
		return nil, &Error{err: err, Scope: licensesTable, Action: action}
	}
	l.ID = (*[32]byte)(id)
	l.Key = (*[32]byte)(key)
	return l, nil
}
