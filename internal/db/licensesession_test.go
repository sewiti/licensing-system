package db

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sewiti/licensing-system/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_InsertLicenseSession(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	ls := &model.LicenseSession{
		ClientID:  base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:  base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey: base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		MachineID: []byte{
			0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
			0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		},
		Created:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Expire:    time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		LicenseID: base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
	}

	mock.ExpectExec("INSERT INTO license_sessions (client_session_id,created,expire,license_id,machine_id,server_session_id,server_session_key) VALUES ($1,$2,$3,$4,$5,$6,$7)").
		WithArgs(
			ls.ClientID[:],
			ls.Created,
			ls.Expire,
			ls.LicenseID[:],
			ls.MachineID,
			ls.ServerID[:],
			ls.ServerKey[:],
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.InsertLicenseSession(context.Background(), ls)
	assert.NoError(t, err)
}

func TestHandler_SelectLicenseSessionByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := &model.LicenseSession{
		ClientID:  base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:  base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey: base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		MachineID: []byte{
			0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
			0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		},
		Created:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Expire:    time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		LicenseID: base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
	}

	rows := sqlmock.NewRows([]string{
		"client_session_id",
		"server_session_id",
		"server_session_key",
		"machine_id",
		"created",
		"expire",
		"license_id",
	}).AddRow(
		expected.ClientID[:],
		expected.ServerID[:],
		expected.ServerKey[:],
		expected.MachineID,
		expected.Created,
		expected.Expire,
		expected.LicenseID[:],
	)

	mock.ExpectQuery("SELECT client_session_id, server_session_id, server_session_key, machine_id, created, expire, license_id FROM license_sessions WHERE client_session_id = $1").
		WithArgs(expected.ClientID[:]).
		WillReturnRows(rows)

	got, err := h.SelectLicenseSessionByID(context.Background(), expected.ClientID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicenseSessionsCountByLicenseID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const expected = 6

	licenseID := base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo=")
	rows := sqlmock.NewRows([]string{"count"}).AddRow(expected)

	mock.ExpectQuery("SELECT COUNT(*) FROM license_sessions WHERE license_id = $1").
		WithArgs(licenseID[:]).
		WillReturnRows(rows)

	got, err := h.SelectLicenseSessionsCountByLicenseID(context.Background(), licenseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_UpdateLicenseSession(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	ls := &model.LicenseSession{
		ClientID:  base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:  base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey: base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		MachineID: []byte{
			0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
			0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		},
		Created:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Expire:    time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		LicenseID: base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
	}

	mock.ExpectExec("UPDATE license_sessions SET created = $1, expire = $2, license_id = $3, machine_id = $4, server_session_id = $5, server_session_key = $6 WHERE client_session_id = $7").
		WithArgs(
			ls.Created,
			ls.Expire,
			ls.LicenseID[:],
			ls.MachineID,
			ls.ServerID[:],
			ls.ServerKey[:],
			ls.ClientID[:],
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.UpdateLicenseSession(context.Background(), ls)
	assert.NoError(t, err)
}

func TestHandler_DeleteLicenseSessionBySessionID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const expected = 4
	clientID := base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y=")

	mock.ExpectExec("DELETE FROM license_sessions WHERE server_session_id = $1").
		WithArgs(clientID[:]).
		WillReturnResult(sqlmock.NewResult(0, expected))

	got, err := h.DeleteLicenseSessionBySessionID(context.Background(), clientID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_DeleteLicenseSessionsByLicenseIDAndMachineID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const expected = 1
	machineID := []byte{
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	}
	licenseID := base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo=")

	mock.ExpectExec("DELETE FROM license_sessions WHERE license_id = $1 AND machine_id = $2").
		WithArgs(licenseID[:], machineID[:]).
		WillReturnResult(sqlmock.NewResult(0, expected))

	got, err := h.DeleteLicenseSessionsByLicenseIDAndMachineID(context.Background(), licenseID, machineID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_DeleteLicenseSessionsExpiredBy(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const expected = 11
	now := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("DELETE FROM license_sessions WHERE expire <= $1").
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, expected))

	got, err := h.DeleteLicenseSessionsExpiredBy(context.Background(), now)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_DeleteLicenseSessionsOverused(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const expected = 54

	mock.ExpectExec("DELETE FROM license_sessions WHERE client_session_id IN (SELECT license_sessions_indexed.client_session_id FROM licenses RIGHT JOIN (SELECT ROW_NUMBER() OVER (PARTITION BY license_id ORDER BY created DESC) AS session_index, client_session_id, license_id FROM license_sessions) AS license_sessions_indexed ON license_sessions_indexed.license_id = licenses.id WHERE license_sessions_indexed.session_index > licenses.max_sessions)").
		WillReturnResult(sqlmock.NewResult(0, expected))

	got, err := h.DeleteLicenseSessionsOverused(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
