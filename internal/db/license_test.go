package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sewiti/licensing-system/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_InsertLicense(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	validUntil := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	l := &model.License{
		ID:          base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
		Key:         base64Key("YFxMq0722e2v2f3tg3+QpkIrV3dlqjCQQv9X7LhMZG0="),
		Note:        "Note",
		Data:        []byte(`{"extraJsonData":true}`),
		MaxSessions: 4,
		ValidUntil:  &validUntil,
		Created:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		IssuerID:    0,
	}

	mock.ExpectExec("INSERT INTO license (created,data,id,issuer_id,key,max_sessions,note,updated,valid_until) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)").
		WithArgs(
			l.Created,
			l.Data,
			l.ID[:],
			l.IssuerID,
			l.Key[:],
			l.MaxSessions,
			l.Note,
			l.Updated,
			l.ValidUntil,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.InsertLicense(context.Background(), l)
	assert.NoError(t, err)
}

func TestHandler_SelectAllLicensesByIssuerID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	validUntil := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	expected := []*model.License{
		{
			ID:          base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
			Key:         base64Key("YFxMq0722e2v2f3tg3+QpkIrV3dlqjCQQv9X7LhMZG0="),
			Note:        "Note",
			Data:        []byte(`{"extraJsonData":true}`),
			MaxSessions: 4,
			ValidUntil:  &validUntil,
			Created:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			Updated:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			IssuerID:    0,
		},
		{
			ID:          base64Key("wf0SXXMDQ03VwgwIIf5TiUO8gT/VzkzihcZ2Z17qomM="),
			Key:         base64Key("7/OninN+j5dqMfQmrQoGkpjTCSdUmLhEHjUarm7qH+Q="),
			Note:        "Note 2",
			Data:        nil,
			MaxSessions: 1,
			ValidUntil:  &validUntil,
			Created:     time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			Updated:     time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			IssuerID:    0,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"key",
		"note",
		"data",
		"max_sessions",
		"valid_until",
		"created",
		"updated",
		"issuer_id",
	})
	for _, v := range expected {
		rows.AddRow(
			v.ID[:],
			v.Key[:],
			v.Note,
			v.Data,
			v.MaxSessions,
			v.ValidUntil,
			v.Created,
			v.Updated,
			v.IssuerID,
		)
	}

	mock.ExpectQuery("SELECT id, key, note, data, max_sessions, valid_until, created, updated, issuer_id FROM license WHERE issuer_id = $1").
		WithArgs(0).
		WillReturnRows(rows)

	got, err := h.SelectAllLicensesByIssuerID(context.Background(), 0)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicenseByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	validUntil := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	expected := &model.License{
		ID:          base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="),
		Key:         base64Key("YFxMq0722e2v2f3tg3+QpkIrV3dlqjCQQv9X7LhMZG0="),
		Note:        "Note",
		Data:        []byte(`{"extraJsonData":true}`),
		MaxSessions: 4,
		ValidUntil:  &validUntil,
		Created:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		IssuerID:    0,
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"key",
		"note",
		"data",
		"max_sessions",
		"valid_until",
		"created",
		"updated",
		"issuer_id",
	}).AddRow(
		expected.ID[:],
		expected.Key[:],
		expected.Note,
		expected.Data,
		expected.MaxSessions,
		expected.ValidUntil,
		expected.Created,
		expected.Updated,
		expected.IssuerID,
	)

	mock.ExpectQuery("SELECT id, key, note, data, max_sessions, valid_until, created, updated, issuer_id FROM license WHERE id = $1").
		WithArgs(expected.ID[:]).
		WillReturnRows(rows)

	got, err := h.SelectLicenseByID(context.Background(), base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="))
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicensesCountByIssuerID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const (
		count    = 12
		issuerID = 0
	)

	rows := sqlmock.NewRows([]string{"count"}).
		AddRow(count)

	mock.ExpectQuery("SELECT COUNT(*) FROM license WHERE issuer_id = $1").
		WithArgs(issuerID).
		WillReturnRows(rows)

	got, err := h.SelectLicensesCountByIssuerID(context.Background(), issuerID)
	assert.NoError(t, err)
	assert.Equal(t, count, got)
}

func TestHandler_UpdateLicense(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	licenseID := base64Key("4k3r5hHKR+PRcaQbjc3yA1cIrZsz3Wixqlv2gouK/y8=")
	update := map[string]interface{}{
		"note":        "new note",
		"data":        json.RawMessage(`{"new":true}`),
		"maxSessions": 2,
		"validUntil":  (*time.Time)(nil),
	}

	mock.ExpectExec("UPDATE license SET data = $1, maxSessions = $2, note = $3, validUntil = $4 WHERE id = $5").
		WithArgs(
			update["data"],
			update["maxSessions"],
			update["note"],
			update["validUntil"],
			licenseID[:],
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.UpdateLicense(context.Background(), licenseID, update)
	assert.NoError(t, err)
}

func TestHandler_DeleteLicenseByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const deleted = 1
	licenseID := base64Key("IgI/tBu0hfqrWiOgNpoyz1gMRfTlBrRiltbecCbTrjY=")

	mock.ExpectExec("DELETE FROM license WHERE id = $1").
		WithArgs(licenseID[:]).
		WillReturnResult(sqlmock.NewResult(0, deleted))

	got, err := h.DeleteLicenseByID(context.Background(), licenseID)
	assert.NoError(t, err)
	assert.Equal(t, deleted, got)
}
