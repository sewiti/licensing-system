package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sewiti/licensing-system/internal/model"
)

// Returns ErrInvalidInput
// Returns SensitiveError
func (c *Core) NewProduct(ctx context.Context, li *model.LicenseIssuer, req *model.Product) (*model.Product, error) {
	if req == nil {
		return nil, fmt.Errorf("%w request", ErrInvalidInput)
	}
	if !ValidProductName(req.Name) {
		return nil, fmt.Errorf("%w name", ErrInvalidInput)
	}
	if req.ContactEmail != "" && !ValidEmail(req.ContactEmail) {
		return nil, fmt.Errorf("%w contact email", ErrInvalidInput)
	}

	now := time.Now()
	p := &model.Product{
		Active:       req.Active,
		Name:         req.Name,
		ContactEmail: req.ContactEmail,
		Data:         req.Data,
		Created:      now,
		Updated:      now,
		IssuerID:     li.ID,
	}
	var err error
	p.ID, err = c.db.InsertProduct(ctx, p)
	return p, handleErrDB(err, "creating product")
}

// Returns SensitiveError
func (c *Core) GetAllProductsByIssuer(ctx context.Context, licenseIssuerID int) ([]*model.Product, error) {
	pp, err := c.db.SelectAllProductsByIssuerID(ctx, licenseIssuerID)
	return pp, handleErrDB(err, "getting all products by issuer")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetProduct(ctx context.Context, productID int) (*model.Product, error) {
	p, err := c.db.SelectProductByID(ctx, productID)
	return p, handleErrDB(err, "getting product")
}

// Returns ErrInvalidInput
// Returns SensitiveError
func (c *Core) UpdateProduct(ctx context.Context, p *model.Product, changes map[string]struct{}) error {
	update := map[string]interface{}{
		"updated": time.Now(),
	}

	if _, ok := changes["active"]; ok {
		update["active"] = p.Active
	}
	if _, ok := changes["name"]; ok {
		if !ValidProductName(p.Name) {
			return fmt.Errorf("%w name", ErrInvalidInput)
		}
		update["name"] = p.Name
	}
	if _, ok := changes["contactEmail"]; ok {
		if p.ContactEmail != "" && !ValidEmail(p.ContactEmail) {
			return fmt.Errorf("%w contact email", ErrInvalidInput)
		}
		update["contact_email"] = p.ContactEmail
	}
	if _, ok := changes["data"]; ok {
		update["data"] = p.Data
	}

	err := c.db.UpdateProduct(ctx, p.ID, update)
	return handleErrDB(err, "updating product")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) DeleteProduct(ctx context.Context, productID, licenseIssuerID int) error {
	_, err := c.db.DeleteProductByID(ctx, productID, licenseIssuerID)
	return handleErrDB(err, "deleting product")
}

func (c *Core) AuthorizeProductUpdate(login *model.LicenseIssuer) (updateMask []string, delete bool) {
	return []string{"active", "name", "contactEmail", "data"}, true
}
