package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const licenseSessionsTable = "license_sessions"

func (h *Handler) InsertLicenseSession(ctx context.Context, ls *model.LicenseSession) error {
	const action = "Insert"
	_, err := h.sq.Insert(licenseSessionsTable).
		SetMap(map[string]interface{}{
			"client_session_id":  ls.ClientID[:],
			"server_session_id":  ls.ServerID[:],
			"server_session_key": ls.ServerKey[:],
			"machine_id":         ls.MachineID,
			"created":            ls.Created,
			"expire":             ls.Expire,
			"license_id":         ls.LicenseID[:],
		}).
		ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: licenseSessionsTable, Action: action}
	}
	return nil
}

func (h *Handler) SelectLicenseSessionByID(ctx context.Context, clientSessionID *[32]byte) (*model.LicenseSession, error) {
	const action = "SelectByID"

	var clientID []byte
	var serverID []byte
	var serverKey []byte
	var licenseID []byte
	row := h.sq.Select(
		"client_session_id",
		"server_session_id",
		"server_session_key",
		"machine_id",
		"created",
		"expire",
		"license_id",
	).
		From(licenseSessionsTable).
		Where(squirrel.Eq{
			"client_session_id": clientSessionID[:],
		}).
		QueryRowContext(ctx)

	ls := &model.LicenseSession{}
	err := h.scanRow(row,
		&clientID,
		&serverID,
		&serverKey,
		&ls.MachineID,
		&ls.Created,
		&ls.Expire,
		&licenseID,
	)
	if err != nil {
		return nil, &Error{err: err, Scope: licenseSessionsTable, Action: action}
	}
	ls.ClientID = (*[32]byte)(clientID)
	ls.ServerID = (*[32]byte)(serverID)
	ls.ServerKey = (*[32]byte)(serverKey)
	ls.LicenseID = (*[32]byte)(licenseID)
	return ls, nil
}

func (h *Handler) SelectLicenseSessionsCountByLicenseID(ctx context.Context, licenseID *[32]byte) (int, error) {
	const (
		action = "SelectCountByLicenseID"
		scope  = licenseSessionsTable
	)
	row := h.sq.Select(
		"COUNT(*)",
	).
		From(licenseSessionsTable).
		Where(squirrel.Eq{
			"license_id": licenseID[:],
		}).
		QueryRowContext(ctx)
	var count int
	err := h.scanRow(row, &count)
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}
	return count, nil
}

func (h *Handler) UpdateLicenseSession(ctx context.Context, ls *model.LicenseSession) error {
	const action = "Update"
	_, err := h.sq.Update(licenseSessionsTable).
		SetMap(map[string]interface{}{
			"server_session_id":  ls.ServerID[:],
			"server_session_key": ls.ServerKey[:],
			"machine_id":         ls.MachineID,
			"created":            ls.Created,
			"expire":             ls.Expire,
			"license_id":         ls.LicenseID[:],
		}).
		Where(squirrel.Eq{
			"client_session_id": ls.ClientID[:],
		}).
		ExecContext(ctx)
	if err != nil {
		return &Error{err: err, Scope: licenseSessionsTable, Action: action}
	}
	return nil
}

func (h *Handler) DeleteLicenseSessionBySessionID(ctx context.Context, clientSessionID *[32]byte) (int, error) {
	sq := h.sq.Delete(licenseSessionsTable).
		Where(squirrel.Eq{
			"license_id": clientSessionID[:],
		})
	return h.execDelete(ctx, sq, licenseSessionsTable, "DeleteBySessionID")
}

func (h *Handler) DeleteLicenseSessionsByLicenseIDAndMachineID(ctx context.Context, licenseID *[32]byte, machineID []byte) (int, error) {
	sq := h.sq.Delete(licenseSessionsTable).
		Where(squirrel.Eq{
			"machine_id": machineID,
			"license_id": licenseID[:],
		})
	return h.execDelete(ctx, sq, licenseSessionsTable, "DeleteByLicenseIDAndMachineID")
}

func (h *Handler) DeleteLicenseSessionsExpiredBy(ctx context.Context, now time.Time) (int, error) {
	sq := h.sq.Delete(licenseSessionsTable).
		Where(squirrel.LtOrEq{
			"expire": now,
		})
	return h.execDelete(ctx, sq, licenseSessionsTable, "DeleteExpiredBy")
}

func (h *Handler) DeleteLicenseSessionsOverused(ctx context.Context) (int, error) {
	const (
		scope  = licenseSessionsTable
		action = "DeleteOverused"
	)

	licenseSessionsIndexed, _, err := h.sq.Select(
		"ROW_NUMBER() OVER (PARTITION BY license_id ORDER BY created DESC) AS session_index",
		"client_session_id",
		"license_id",
	).
		From(licenseSessionsTable).
		ToSql()
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}

	overused, _, err := h.sq.Select("license_sessions_indexed.client_session_id").
		From(licensesTable).
		RightJoin(
			fmt.Sprintf("(%s) AS %s ON %s",
				licenseSessionsIndexed,
				"license_sessions_indexed",
				"license_sessions_indexed.license_id = licenses.id",
			),
		).
		Where("license_sessions_indexed.session_index > licenses.max_sessions").
		ToSql()
	if err != nil {
		return 0, &Error{err: err, Scope: scope, Action: action}
	}

	sq := h.sq.Delete(licenseSessionsTable).
		Where(fmt.Sprintf("client_session_id IN (%s)", overused))

	return h.execDelete(ctx, sq, scope, action)
}
