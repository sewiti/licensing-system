package core

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sewiti/licensing-system/internal/model"
	"golang.org/x/crypto/nacl/box"
)

var licenseRemap = map[string]string{
	"maxSessions": "max_sessions",
	"validUntil":  "valid_until",
}

// Returns ErrInvalidInput
// Returns ErrExceedsLimit
// Returns SensitiveError
func (c *Core) NewLicense(ctx context.Context, li *model.LicenseIssuer, note string, data json.RawMessage, maxSessions int, validUntil *time.Time) (*model.License, error) {
	if !ValidNote(note) {
		return nil, fmt.Errorf("%w note", ErrInvalidInput)
	}
	if maxSessions <= 0 {
		return nil, fmt.Errorf("%w max sessions", ErrInvalidInput)
	}
	count, err := c.db.SelectLicensesCountByIssuerID(ctx, li.ID)
	if err != nil {
		return nil, handleErrDB(err, "counting licenses")
	}
	if !li.MaxLicenses.Allows(count + 1) {
		return nil, fmt.Errorf("max sessions: %w", ErrExceedsLimit)
	}

	id, key, err := box.GenerateKey(cryptorand.Reader)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	l := &model.License{
		ID:          id,
		Key:         key,
		Note:        note,
		Data:        data,
		MaxSessions: maxSessions,
		ValidUntil:  validUntil,
		Created:     now,
		Updated:     now,
		IssuerID:    li.ID,
	}
	err = c.db.InsertLicense(ctx, l)
	return l, handleErrDB(err, "creating license")
}

// Returns SensitiveError
func (c *Core) GetAllLicensesByIssuer(ctx context.Context, licenseIssuerID int) ([]*model.License, error) {
	ll, err := c.db.SelectAllLicensesByIssuerID(ctx, licenseIssuerID)
	return ll, handleErrDB(err, "getting all licenses")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetLicense(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	l, err := c.db.SelectLicenseByID(ctx, licenseID)
	return l, handleErrDB(err, "getting license")
}

// Returns ErrInvalidInput
// Returns SensitiveError
func (c *Core) UpdateLicense(ctx context.Context, licenseID *[32]byte, update map[string]interface{}) error {
	// Validate note
	v, ok := update["note"]
	if ok {
		note, ok := v.(string)
		if !ok || !ValidNote(note) {
			return fmt.Errorf("%w note", ErrInvalidInput)
		}
	}
	// Validate max sessions
	v, ok = update["maxSessions"]
	if ok {
		maxSession, ok := v.(int)
		if !ok || maxSession <= 0 {
			return fmt.Errorf("%w max sessions", ErrInvalidInput)
		}
	}

	updateApplyRemap(update, licenseRemap) // Remap json keys to db keys
	update["updated"] = time.Now()
	err := c.db.UpdateLicense(ctx, licenseID, update)
	return handleErrDB(err, "updating license")
}

// Returns SensitiveError
func (c *Core) DeleteLicense(ctx context.Context, licenseID *[32]byte) error {
	_, err := c.db.DeleteLicenseByID(ctx, licenseID)
	return handleErrDB(err, "deleting license")
}

func (c *Core) AuthorizeLicenseUpdate(login *model.LicenseIssuer) (updateMask []string, delete bool) {
	return []string{"note", "data", "maxSessions", "validUntil"}, true
}
