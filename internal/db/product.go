package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/sewiti/licensing-system/internal/model"
)

const productTable = "product"

func (h *Handler) InsertProduct(ctx context.Context, p *model.Product) (int, error) {
	const (
		action = "Insert"
		scope  = productTable
	)
	sq := h.sq.Insert(scope).
		SetMap(map[string]interface{}{
			"active":        p.Active,
			"name":          p.Name,
			"contact_email": p.ContactEmail,
			"data":          p.Data,
			"created":       p.Created,
			"updated":       p.Updated,
			"issuer_id":     p.IssuerID,
		}).Suffix("RETURNING id")

	var id int
	return id, h.execInsert(ctx, sq, scope, action, &id)
}

func (h *Handler) SelectAllProductsByIssuerID(ctx context.Context, licenseIssuerID int) ([]*model.Product, error) {
	return h.selectProducts(ctx, "SelectAllByIssuerID", func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
		return sq.Where(squirrel.Eq{
			"issuer_id": licenseIssuerID,
		}).OrderBy("active DESC", "id")
	})
}

func (h *Handler) SelectProductByID(ctx context.Context, productID int) (*model.Product, error) {
	return h.selectProduct(ctx, "SelectByID",
		func(sq squirrel.SelectBuilder) squirrel.SelectBuilder {
			return sq.Where(squirrel.Eq{
				"id": productID,
			})
		})
}

func (h *Handler) selectProduct(ctx context.Context, action string, d selectDecorator) (*model.Product, error) {
	pp, err := h.selectProducts(ctx, action, d)
	if err != nil {
		return nil, err
	}
	if len(pp) == 0 {
		return nil, &Error{err: ErrNotFound, Scope: productTable, Action: action}
	}
	return pp[0], nil
}

func (h *Handler) selectProducts(ctx context.Context, action string, d selectDecorator) ([]*model.Product, error) {
	const scope = productTable

	sq := h.sq.Select(
		"id",
		"active",
		"name",
		"contact_email",
		"data",
		"created",
		"updated",
		"issuer_id",
	).From(scope)

	rows, err := d(sq).QueryContext(ctx)
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	defer rows.Close()

	var pp []*model.Product
	for rows.Next() {
		p := &model.Product{}
		err = rows.Scan(
			&p.ID,
			&p.Active,
			&p.Name,
			&p.ContactEmail,
			&p.Data,
			&p.Created,
			&p.Updated,
			&p.IssuerID,
		)
		if err != nil {
			return nil, &Error{err: err, Scope: scope, Action: action}
		}
		pp = append(pp, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, &Error{err: err, Scope: scope, Action: action}
	}
	return pp, nil
}

func (h *Handler) UpdateProduct(ctx context.Context, productID int, update map[string]interface{}) error {
	const (
		action = "Update"
		scope  = productTable
	)
	sq := h.sq.Update(scope).
		SetMap(update).
		Where(squirrel.Eq{
			"id": productID,
		})
	return h.execUpdate(ctx, sq, scope, action)
}

func (h *Handler) DeleteProductByID(ctx context.Context, productID, licenseIssuerID int) (int, error) {
	const scope = productTable
	sq := h.sq.Delete(scope).
		Where(squirrel.Eq{
			"id":        productID,
			"issuer_id": licenseIssuerID,
		})
	return h.execDelete(ctx, sq, scope, "DeleteByID")
}
