package db

import (
	"context"
	"encoding/json"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/pkg/model"
)

const licensesTable = "licenses"

func (h *Handler) InsertLicense(ctx context.Context, l *model.License) error {
	const action = "Insert"
	data, err := json.Marshal(l.Data)
	if err != nil {
		return &Error{err: err, Scope: licensesTable, Action: action}
	}
	_, err = h.sq.Insert(licensesTable).
		SetMap(map[string]interface{}{
			"id":           l.ID[:],
			"key":          l.Key[:],
			"note":         l.Note,
			"data":         data,
			"max_sessions": l.MaxSessions,
			"valid_until":  l.ValidUntil,
			"created":      l.Created,
			"updated":      l.Updated,
			"issuer_id":    l.IssuerID,
		}).
		ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: licensesTable, Action: action}
	}
	return nil
}

func (h *Handler) SelectLicenseByID(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	const action = "SelectByID"
	row := h.sq.Select(
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
		}).
		QueryRowContext(ctx)

	var id []byte
	var key []byte
	var data []byte
	l := &model.License{}
	err := row.Scan(
		&id,
		&key,
		&l.Note,
		&data,
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
	err = json.Unmarshal(data, &l.Data)
	if err != nil {
		return nil, &Error{err: err, Scope: licensesTable, Action: action}
	}
	return l, nil
}
