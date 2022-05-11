package core

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sewiti/licensing-system/internal/model"
	"github.com/sewiti/licensing-system/pkg/util"
)

// Returns ErrInvalidInput
// Returns ErrExceedsLimit
// Returns SensitiveError
func (c *Core) NewLicense(ctx context.Context, li *model.LicenseIssuer, req *model.License) (*model.License, error) {
	if req == nil {
		return nil, fmt.Errorf("%w request", ErrInvalidInput)
	}
	if !ValidLicenseName(req.Name) {
		return nil, fmt.Errorf("%w name", ErrInvalidInput)
	}
	if !ValidLicenseTags(req.Tags) {
		return nil, fmt.Errorf("%w tags", ErrInvalidInput)
	}
	if req.EndUserEmail != "" && !ValidEmail(req.EndUserEmail) {
		return nil, fmt.Errorf("%w end user email", ErrInvalidInput)
	}
	if !ValidLicenseNote(req.Note) {
		return nil, fmt.Errorf("%w note", ErrInvalidInput)
	}
	if req.MaxSessions <= 0 {
		return nil, fmt.Errorf("%w max sessions", ErrInvalidInput)
	}
	count, err := c.db.SelectLicensesCountByIssuerID(ctx, li.ID)
	if err != nil {
		return nil, handleErrDB(err, "counting licenses")
	}
	if !li.MaxLicenses.Allows(count + 1) {
		return nil, fmt.Errorf("max licenses: %w", ErrExceedsLimit)
	}

	id, key, err := util.GenerateKey(cryptorand.Reader)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	l := &model.License{
		ID:           id,
		Key:          key,
		Active:       req.Active,
		Name:         req.Name,
		Tags:         req.Tags,
		EndUserEmail: req.EndUserEmail,
		Note:         req.Note,
		Data:         req.Data,
		MaxSessions:  req.MaxSessions,
		ValidUntil:   req.ValidUntil,
		Created:      now,
		Updated:      now,
		LastUsed:     nil,
		IssuerID:     li.ID,
		ProductID:    req.ProductID,
	}
	err = c.db.InsertLicense(ctx, l)
	return l, handleErrDB(err, "creating license")
}

// Returns SensitiveError
func (c *Core) GetAllLicensesByIssuer(ctx context.Context, licenseIssuerID int) ([]*model.License, error) {
	ll, err := c.db.SelectAllLicensesByIssuerID(ctx, licenseIssuerID)
	return ll, handleErrDB(err, "getting all licenses by issuer")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetLicense(ctx context.Context, licenseID []byte) (*model.License, error) {
	l, err := c.db.SelectLicenseByID(ctx, licenseID)
	return l, handleErrDB(err, "getting license")
}

// Returns ErrInvalidInput
// Returns SensitiveError
func (c *Core) UpdateLicense(ctx context.Context, l *model.License, changes map[string]struct{}) error {
	update := map[string]interface{}{
		"updated": time.Now(),
	}

	if _, ok := changes["active"]; ok {
		update["active"] = l.Active
	}
	if _, ok := changes["name"]; ok {
		if !ValidLicenseName(l.Name) {
			return fmt.Errorf("%w name", ErrInvalidInput)
		}
		update["name"] = l.Name
	}
	if _, ok := changes["tags"]; ok {
		if !ValidLicenseTags(l.Tags) {
			return fmt.Errorf("%w tags", ErrInvalidInput)
		}
		update["tags"] = pq.Array(l.Tags)
	}
	if _, ok := changes["endUserEmail"]; ok {
		if l.EndUserEmail != "" && !ValidEmail(l.EndUserEmail) {
			return fmt.Errorf("%w end user email", ErrInvalidInput)
		}
		update["end_user_email"] = l.EndUserEmail
	}
	if _, ok := changes["note"]; ok {
		if !ValidLicenseNote(l.Note) {
			return fmt.Errorf("%w note", ErrInvalidInput)
		}
		update["note"] = l.Note
	}
	if _, ok := changes["data"]; ok {
		update["data"] = l.Data
	}
	if _, ok := changes["maxSessions"]; ok {
		if l.MaxSessions <= 0 {
			return fmt.Errorf("%w max sessions", ErrInvalidInput)
		}
		update["max_sessions"] = l.MaxSessions
	}
	if _, ok := changes["validUntil"]; ok {
		update["valid_until"] = l.ValidUntil
	}
	if _, ok := changes["lastUsed"]; ok {
		update["last_used"] = l.LastUsed
	}

	err := c.db.UpdateLicense(ctx, l.ID, l.IssuerID, update)
	return handleErrDB(err, "updating license")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) DeleteLicense(ctx context.Context, licenseID []byte, licenseIssuerID int) error {
	_, err := c.db.DeleteLicenseByID(ctx, licenseID, licenseIssuerID)
	return handleErrDB(err, "deleting license")
}

func (c *Core) AuthorizeLicenseUpdate(login *model.LicenseIssuer) (updateMask []string, delete bool) {
	return []string{"active", "name", "tags", "endUserEmail", "note", "data", "maxSessions", "validUntil", "productID"}, true
}
