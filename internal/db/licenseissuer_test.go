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

func TestHandler_InsertLicenseIssuer(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	li := &model.LicenseIssuer{
		ID:           4,
		Active:       true,
		Username:     "superadmin",
		PasswordHash: "argon2id$t=2,m=65536,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$MWGxLaIDexqiJjL1OjDFA8x+stulQAzkN6g65n9ugGs=",
		MaxLicenses:  -1,
		Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec("INSERT INTO license_issuer (active,created,max_licenses,password_hash,updated,username) VALUES ($1,$2,$3,$4,$5,$6)").
		WithArgs(
			li.Active,
			li.Created,
			li.MaxLicenses,
			li.PasswordHash,
			li.Updated,
			li.Username,
		).
		WillReturnResult(sqlmock.NewResult(int64(li.ID), 1))

	id, err := h.InsertLicenseIssuer(context.Background(), li)
	assert.NoError(t, err)
	assert.Equal(t, li.ID, id)
}

func TestHandler_SelectAllLicenseIssuers(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := []*model.LicenseIssuer{
		{
			ID:           0,
			Active:       true,
			Username:     "superadmin",
			PasswordHash: "argon2id$v=19,t=2,m=65536,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$MWGxLaIDexqiJjL1OjDFA8x+stulQAzkN6g65n9ugGs=",
			MaxLicenses:  -1,
			Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           1,
			Active:       true,
			Username:     "testuser",
			PasswordHash: "argon2id$v=19,t=8,m=32768,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$JMhcCrhKMFOtY5rcEmnAiDKw71ooKOGwIaeermvmouw=",
			MaxLicenses:  -1,
			Created:      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			Updated:      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"active",
		"username",
		"password_hash",
		"max_licenses",
		"created",
		"updated",
	})
	for _, v := range expected {
		rows.AddRow(
			v.ID,
			v.Active,
			v.Username,
			v.PasswordHash,
			v.MaxLicenses,
			v.Created,
			v.Updated,
		)
	}

	mock.ExpectQuery("SELECT id, active, username, password_hash, max_licenses, created, updated FROM license_issuer").
		WillReturnRows(rows)

	got, err := h.SelectAllLicenseIssuers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicenseIssuerByUsername(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := &model.LicenseIssuer{
		ID:           0,
		Active:       true,
		Username:     "superadmin",
		PasswordHash: "argon2id$t=2,m=65536,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$MWGxLaIDexqiJjL1OjDFA8x+stulQAzkN6g65n9ugGs=",
		MaxLicenses:  -1,
		Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"active",
		"username",
		"password_hash",
		"max_licenses",
		"created",
		"updated",
	}).AddRow(
		expected.ID,
		expected.Active,
		expected.Username,
		expected.PasswordHash,
		expected.MaxLicenses,
		expected.Created,
		expected.Updated,
	)

	mock.ExpectQuery("SELECT id, active, username, password_hash, max_licenses, created, updated FROM license_issuer WHERE username = $1").
		WithArgs(expected.Username).
		WillReturnRows(rows)

	got, err := h.SelectLicenseIssuerByUsername(context.Background(), expected.Username)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectLicenseIssuerByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := &model.LicenseIssuer{
		ID:           0,
		Active:       true,
		Username:     "superadmin",
		PasswordHash: "argon2id$t=2,m=65536,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$MWGxLaIDexqiJjL1OjDFA8x+stulQAzkN6g65n9ugGs=",
		MaxLicenses:  -1,
		Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"active",
		"username",
		"password_hash",
		"max_licenses",
		"created",
		"updated",
	}).AddRow(
		expected.ID,
		expected.Active,
		expected.Username,
		expected.PasswordHash,
		expected.MaxLicenses,
		expected.Created,
		expected.Updated,
	)

	mock.ExpectQuery("SELECT id, active, username, password_hash, max_licenses, created, updated FROM license_issuer WHERE id = $1").
		WithArgs(expected.ID).
		WillReturnRows(rows)

	got, err := h.SelectLicenseIssuerByID(context.Background(), expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_UpdateLicenseIssuer(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const licenseIssuerID = 3
	update := map[string]interface{}{
		"active":       true,
		"username":     "test-username",
		"max_licenses": 3,
		"created":      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		"updated":      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec("UPDATE license_issuer SET active = $1, created = $2, max_licenses = $3, updated = $4, username = $5 WHERE id = $6").
		WithArgs(
			update["active"],
			update["created"],
			update["max_licenses"],
			update["updated"],
			update["username"],
			licenseIssuerID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.UpdateLicenseIssuer(context.Background(), licenseIssuerID, update)
	assert.NoError(t, err)
}

func TestHandler_UpdateLicenseIssuerByUsername(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const username = "test-user"
	update := map[string]interface{}{
		"active":       true,
		"username":     "test-username",
		"max_licenses": 3,
		"created":      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		"updated":      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec("UPDATE license_issuer SET active = $1, created = $2, max_licenses = $3, updated = $4, username = $5 WHERE username = $6").
		WithArgs(
			update["active"],
			update["created"],
			update["max_licenses"],
			update["updated"],
			update["username"],
			username,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.UpdateLicenseIssuerByUsername(context.Background(), username, update)
	assert.NoError(t, err)
}

func TestHandler_DeleteLicenseIssuerByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const (
		deleted         = 1
		licenseIssuerID = 69
	)

	mock.ExpectExec("DELETE FROM license_issuer WHERE id = $1").
		WithArgs(licenseIssuerID).
		WillReturnResult(sqlmock.NewResult(0, deleted))

	got, err := h.DeleteLicenseIssuerByID(context.Background(), licenseIssuerID)
	assert.NoError(t, err)
	assert.Equal(t, deleted, got)
}
