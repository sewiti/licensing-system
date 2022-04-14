package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const licenseSessionTable = "license_session"

func (h *Handler) InsertLicenseSession(ctx context.Context, ls *model.LicenseSession) error {
	const (
		action = "Insert"
		scope  = licenseSessionTable
	)
	sq := h.sq.Insert(scope).
		SetMap(map[string]interface{}{
			"client_session_id":  ls.ClientID,
			"server_session_id":  ls.ServerID,
			"server_session_key": ls.ServerKey,
			"identifier":         ls.Identifier,
			"machine_id":         ls.MachineID,
			"app_version":        ls.AppVersion,
			"created":            ls.Created,
			"expire":             ls.Expire,
			"license_id":         ls.LicenseID,
		})
	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) SelectAllLicenseSessionsByLicenseID(ctx context.Context, licenseID []byte) ([]*model.LicenseSession, error) {
	return h.selectLicenseSessions(ctx, "SelectAllByLicenseID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"license_id": licenseID,
			}).OrderBy("created")
		})
}

func (h *Handler) SelectLicenseSessionByID(ctx context.Context, clientSessionID []byte) (*model.LicenseSession, error) {
	return h.selectLicenseSession(ctx, "SelectByID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"client_session_id": clientSessionID,
			})
		})
}

func (h *Handler) selectLicenseSession(ctx context.Context, action string, d selectDecorator) (*model.LicenseSession, error) {
	lss, err := h.selectLicenseSessions(ctx, action, d)
	if err != nil {
		return nil, err
	}
	if len(lss) == 0 {
		return nil, &Error{err: ErrNotFound, Scope: licenseSessionTable, Action: action}
	}
	return lss[0], nil
}

func (h *Handler) selectLicenseSessions(ctx context.Context, action string, d selectDecorator) ([]*model.LicenseSession, error) {
	const scope = licenseSessionTable

	sq := h.sq.Select(
		"client_session_id",
		"server_session_id",
		"server_session_key",
		"identifier",
		"machine_id",
		"app_version",
		"created",
		"expire",
		"license_id",
	).From(scope)

	rows, err := d(sq).QueryContext(ctx)
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	defer rows.Close()

	var lss []*model.LicenseSession
	for rows.Next() {
		ls := &model.LicenseSession{}
		err = rows.Scan(
			&ls.ClientID,
			&ls.ServerID,
			&ls.ServerKey,
			&ls.Identifier,
			&ls.MachineID,
			&ls.AppVersion,
			&ls.Created,
			&ls.Expire,
			&ls.LicenseID,
		)
		if err != nil {
			return nil, &Error{err: err, Scope: scope, Action: action}
		}
		lss = append(lss, ls)
	}

	err = rows.Err()
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	return lss, nil
}

func (h *Handler) UpdateLicenseSession(ctx context.Context, ls *model.LicenseSession) error {
	const (
		action = "Update"
		scope  = licenseSessionTable
	)
	sq := h.sq.Update(scope).
		SetMap(map[string]interface{}{
			"server_session_id":  ls.ServerID,
			"server_session_key": ls.ServerKey,
			"identifier":         ls.Identifier,
			"machine_id":         ls.MachineID,
			"app_version":        ls.AppVersion,
			"created":            ls.Created,
			"expire":             ls.Expire,
			"license_id":         ls.LicenseID,
		}).
		Where(squirrel.Eq{
			"client_session_id": ls.ClientID,
		})

	_, err := sq.ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: scope, Action: action}
	}
	return nil
}

func (h *Handler) DeleteLicenseSessionBySessionID(ctx context.Context, clientSessionID []byte) (int, error) {
	sq := h.sq.Delete(licenseSessionTable).
		Where(squirrel.Eq{
			"client_session_id": clientSessionID,
		})
	return h.execDelete(ctx, sq, licenseSessionTable, "DeleteBySessionID")
}

func (h *Handler) DeleteLicenseSessionsByLicenseIDAndMachineID(ctx context.Context, licenseID []byte, machineID []byte) (int, error) {
	sq := h.sq.Delete(licenseSessionTable).
		Where(squirrel.Eq{
			"machine_id": machineID,
			"license_id": licenseID,
		})
	return h.execDelete(ctx, sq, licenseSessionTable, "DeleteByLicenseIDAndMachineID")
}

func (h *Handler) DeleteLicenseSessionsExpiredBy(ctx context.Context, now time.Time) (int, error) {
	sq := h.sq.Delete(licenseSessionTable).
		Where(squirrel.LtOrEq{
			"expire": now,
		})
	return h.execDelete(ctx, sq, licenseSessionTable, "DeleteExpiredBy")
}

func (h *Handler) DeleteLicenseSessionsOverused(ctx context.Context) (int, error) {
	const (
		scope  = licenseSessionTable
		action = "DeleteOverused"
	)

	licenseSessionsIndexed, _, err := h.sq.Select(
		"ROW_NUMBER() OVER (PARTITION BY license_id ORDER BY created DESC) AS session_index",
		"client_session_id",
		"license_id",
	).
		From(licenseSessionTable).
		ToSql()
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}

	overused, _, err := h.sq.Select("license_session_indexed.client_session_id").
		From(licenseTable).
		RightJoin(
			fmt.Sprintf("(%s) AS %s ON %s",
				licenseSessionsIndexed,
				"license_session_indexed",
				"license_session_indexed.license_id = license.id",
			),
		).
		Where("license_session_indexed.session_index > license.max_sessions").
		ToSql()
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}

	sq := h.sq.Delete(licenseSessionTable).
		Where(fmt.Sprintf("client_session_id IN (%s)", overused))

	return h.execDelete(ctx, sq, scope, action)
}
