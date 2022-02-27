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

	mock.ExpectExec("INSERT INTO licenses (created,data,id,issuer_id,key,max_sessions,note,updated,valid_until) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)").
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

	mock.ExpectQuery("SELECT id, key, note, data, max_sessions, valid_until, created, updated, issuer_id FROM licenses WHERE id = $1").
		WithArgs(expected.ID[:]).
		WillReturnRows(rows)

	got, err := h.SelectLicenseByID(context.Background(), base64Key("sswRe+P3j0nKqTcCLJ+cPk/8VyjrJzNyxcHCUoXYDFo="))
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
