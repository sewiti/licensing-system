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
		ClientID:   base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:   base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey:  base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		Identifier: "licensing | Linux 5.10.0-11-amd64 x86_64 | #1 SMP Debian 5.10.92-1 (2022-01-18)",
		MachineID: []byte{
			0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
			0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		},
		AppVersion: "1.42",
		Created:    time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Expire:     time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		LicenseID:  base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
	}

	mock.ExpectExec("INSERT INTO license_session (app_version,client_session_id,created,expire,identifier,license_id,machine_id,server_session_id,server_session_key) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)").
		WithArgs(
			ls.AppVersion,
			ls.ClientID,
			ls.Created,
			ls.Expire,
			ls.Identifier,
			ls.LicenseID,
			ls.MachineID,
			ls.ServerID,
			ls.ServerKey,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.InsertLicenseSession(context.Background(), ls)
	assert.NoError(t, err)
}

func TestHandler_SelectAllLicenseSessionsByLicenseID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	licenseID := base64Key("rnlnMc3JAaIPfxNnYv/A7WT+QpzUuFs3h6pali4V8T4=")
	expected := []*model.LicenseSession{
		{
			AppVersion: "1.23",
			ClientID:   base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
			ServerID:   base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
			ServerKey:  base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
			Identifier: "licensing | Linux 5.10.0-11-amd64 x86_64 | #1 SMP Debian 5.10.92-1 (2022-01-18)",
			MachineID: []byte{
				0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
				0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
			},
			Created:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			Expire:    time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			LicenseID: licenseID,
		},
		{
			AppVersion: "1.62",
			ClientID:   base64Key("9IdvR71TDTcV0aYS9EyrTU09tzM9+LqlaGb5cXwiPQU="),
			ServerID:   base64Key("kciqYBC47Y03z7umzyQu6LoAmOu6xkoyOwOMNPeXD1M="),
			ServerKey:  base64Key("vj708c0gEF97Z/GiOrsEcAbHLHM5BJdSRZpg5YyxbTw="),
			Identifier: "licensing-2 | Linux 5.10.0-11-amd64 x86_64 | #1 SMP Debian 5.10.92-1 (2022-01-18)",
			MachineID: []byte{
				0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
				0xf, 0xe, 0xd, 0xc, 0xb, 0xa, 0x9, 0x8,
			},
			Created:   time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			Expire:    time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC),
			LicenseID: licenseID,
		},
	}

	rows := sqlmock.NewRows([]string{
		"client_session_id",
		"server_session_id",
		"server_session_key",
		"identifier",
		"machine_id",
		"app_version",
		"created",
		"expire",
		"license_id",
	})
	for _, v := range expected {
		rows.AddRow(
			v.ClientID,
			v.ServerID,
			v.ServerKey,
			v.Identifier,
			v.MachineID,
			v.AppVersion,
			v.Created,
			v.Expire,
			v.LicenseID,
		)
	}

	mock.ExpectQuery("SELECT client_session_id, server_session_id, server_session_key, identifier, machine_id, app_version, created, expire, license_id FROM license_session WHERE license_id = $1 ORDER BY created").
		WithArgs(licenseID).
		WillReturnRows(rows)

	got, err := h.SelectAllLicenseSessionsByLicenseID(context.Background(), licenseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicenseSessionByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := &model.LicenseSession{
		AppVersion: "1.42",
		ClientID:   base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:   base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey:  base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		Identifier: "licensing | Linux 5.10.0-11-amd64 x86_64 | #1 SMP Debian 5.10.92-1 (2022-01-18)",
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
		"identifier",
		"machine_id",
		"app_version",
		"created",
		"expire",
		"license_id",
	}).AddRow(
		expected.ClientID,
		expected.ServerID,
		expected.ServerKey,
		expected.Identifier,
		expected.MachineID,
		expected.AppVersion,
		expected.Created,
		expected.Expire,
		expected.LicenseID,
	)

	mock.ExpectQuery("SELECT client_session_id, server_session_id, server_session_key, identifier, machine_id, app_version, created, expire, license_id FROM license_session WHERE client_session_id = $1").
		WithArgs(expected.ClientID).
		WillReturnRows(rows)

	got, err := h.SelectLicenseSessionByID(context.Background(), expected.ClientID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_UpdateLicenseSession(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	ls := &model.LicenseSession{
		AppVersion: "1.23",
		ClientID:   base64Key("TK3hUQGPZKiqXGpG76D9VGrjfvqjDXisv7nB7Qgm20Y="),
		ServerID:   base64Key("XyVhUg+vvJ6Z4RtCDEyW25OSxDSeySDvVzMHr1iGfwc="),
		ServerKey:  base64Key("omTAEtJDlu4+o+1xwhEujmpDv94+ljwDKydf2mjYL0A="),
		Identifier: "licensing | Linux 5.10.0-11-amd64 x86_64 | #1 SMP Debian 5.10.92-1 (2022-01-18)",
		MachineID: []byte{
			0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
			0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		},
		Created:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Expire:    time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		LicenseID: base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
	}

	mock.ExpectExec("UPDATE license_session SET app_version = $1, created = $2, expire = $3, identifier = $4, license_id = $5, machine_id = $6, server_session_id = $7, server_session_key = $8 WHERE client_session_id = $9").
		WithArgs(
			ls.AppVersion,
			ls.Created,
			ls.Expire,
			ls.Identifier,
			ls.LicenseID,
			ls.MachineID,
			ls.ServerID,
			ls.ServerKey,
			ls.ClientID,
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

	mock.ExpectExec("DELETE FROM license_session WHERE client_session_id = $1").
		WithArgs(clientID).
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

	mock.ExpectExec("DELETE FROM license_session WHERE license_id = $1 AND machine_id = $2").
		WithArgs(licenseID, machineID).
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

	mock.ExpectExec("DELETE FROM license_session WHERE expire <= $1").
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

	mock.ExpectExec(
		"DELETE FROM license_session WHERE client_session_id IN " +
			"(SELECT license_session_indexed.client_session_id FROM license RIGHT JOIN " +
			"(SELECT ROW_NUMBER() OVER (PARTITION BY license_id ORDER BY created DESC) AS session_index, client_session_id, license_id " +
			"FROM license_session) AS license_session_indexed ON license_session_indexed.license_id = license.id " +
			"WHERE license_session_indexed.session_index > license.max_sessions)",
	).
		WillReturnResult(sqlmock.NewResult(0, expected))

	got, err := h.DeleteLicenseSessionsOverused(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
