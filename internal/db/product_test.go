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

func TestHandler_InsertProduct(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	p := &model.Product{
		ID:           4,
		Active:       true,
		Name:         "john",
		ContactEmail: "john@email.com",
		Data:         []byte(`{"hello":"world!"}`),
		IssuerID:     3,
		Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectQuery("INSERT INTO product (active,contact_email,created,data,issuer_id,name,updated) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id").
		WithArgs(
			p.Active,
			p.ContactEmail,
			p.Created,
			p.Data,
			p.IssuerID,
			p.Name,
			p.Updated,
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(p.ID))

	id, err := h.InsertProduct(context.Background(), p)
	assert.NoError(t, err)
	assert.Equal(t, p.ID, id)
}

func TestHandler_SelectAllProductsByIssuerID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	issuerID := 5
	expected := []*model.Product{
		{
			ID:           0,
			Active:       true,
			Name:         "superadmin",
			ContactEmail: "email@test.com",
			Data:         []byte(`hello`),
			Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			IssuerID:     issuerID,
		},
		{
			ID:           1,
			Active:       true,
			Name:         "testuser",
			ContactEmail: "email@test.com",
			Data:         []byte("+370123123123"),
			Created:      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			Updated:      time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			IssuerID:     issuerID,
		},
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"active",
		"name",
		"contact_email",
		"data",
		"created",
		"updated",
		"issuer_id",
	})
	for _, v := range expected {
		rows.AddRow(
			v.ID,
			v.Active,
			v.Name,
			v.ContactEmail,
			v.Data,
			v.Created,
			v.Updated,
			v.IssuerID,
		)
	}

	mock.ExpectQuery("SELECT id, active, name, contact_email, data, created, updated, issuer_id FROM product WHERE issuer_id = $1 ORDER BY active DESC, id").
		WillReturnRows(rows)

	got, err := h.SelectAllProductsByIssuerID(context.Background(), issuerID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_SelectProductByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	expected := &model.Product{
		ID:           0,
		Active:       true,
		Name:         "superadmin",
		ContactEmail: "email@test.com",
		Data:         []byte("+370123123123"),
		Created:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		IssuerID:     5,
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"active",
		"name",
		"contact_email",
		"data",
		"created",
		"updated",
		"issuer_id",
	}).AddRow(
		expected.ID,
		expected.Active,
		expected.Name,
		expected.ContactEmail,
		expected.Data,
		expected.Created,
		expected.Updated,
		expected.IssuerID,
	)

	mock.ExpectQuery("SELECT id, active, name, contact_email, data, created, updated, issuer_id FROM product WHERE id = $1").
		WithArgs(expected.ID).
		WillReturnRows(rows)

	got, err := h.SelectProductByID(context.Background(), expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestHandler_UpdateProduct(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const productID = 3
	update := map[string]interface{}{
		"active":        true,
		"name":          "test-username",
		"contact_email": "email@test.com",
		"created":       time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		"updated":       time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec("UPDATE product SET active = $1, contact_email = $2, created = $3, name = $4, updated = $5 WHERE id = $6").
		WithArgs(
			update["active"],
			update["contact_email"],
			update["created"],
			update["name"],
			update["updated"],
			productID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = h.UpdateProduct(context.Background(), productID, update)
	assert.NoError(t, err)
}

func TestHandler_DeleteProductByID(t *testing.T) {
	h, mock, err := newMock()
	require.NoError(t, err)
	defer h.Close()

	const (
		deleted         = 1
		licenseIssuerID = 3
		productID       = 69
	)

	mock.ExpectExec("DELETE FROM product WHERE id = $1 AND issuer_id = $2").
		WithArgs(productID, licenseIssuerID).
		WillReturnResult(sqlmock.NewResult(0, deleted))

	got, err := h.DeleteProductByID(context.Background(), productID, licenseIssuerID)
	assert.NoError(t, err)
	assert.Equal(t, deleted, got)
}
