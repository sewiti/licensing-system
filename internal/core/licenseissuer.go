package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sewiti/licensing-system/internal/core/auth"
	"github.com/sewiti/licensing-system/internal/model"
)

func CLILogin() *model.LicenseIssuer {
	return &model.LicenseIssuer{ID: -1}
}

var licenseIssuerRemap = map[string]string{
	"maxLicenses": "max_licenses",
}

// Returns ErrInvalidInput
// Returns ErrPasswdTooWeak
// Returns ErrDuplicate
// Returns SensitiveError
func (c *Core) NewLicenseIssuer(ctx context.Context, username, password string, maxLicenses model.Limit) (*model.LicenseIssuer, error) {
	if !ValidUsername(username) {
		return nil, fmt.Errorf("%w username", ErrInvalidInput)
	}
	entropy, ok := c.SufficientPasswdStrength(username, password)
	if !ok {
		return nil, fmt.Errorf("%w: entropy %.2f", ErrPasswdTooWeak, entropy)
	}
	passwdHash, err := auth.HashPasswd(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	li := &model.LicenseIssuer{
		Active:       true,
		Username:     username,
		PasswordHash: passwdHash,
		MaxLicenses:  maxLicenses,
		Created:      now,
		Updated:      now,
	}

	li.ID, err = c.db.InsertLicenseIssuer(ctx, li)
	return li, handleErrDB(err, "creating license issuer")
}

// Returns SensitiveError
func (c *Core) GetAllLicenseIssuers(ctx context.Context) ([]*model.LicenseIssuer, error) {
	lii, err := c.db.SelectAllLicenseIssuers(ctx)
	return lii, handleErrDB(err, "getting all license issuers")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetLicenseIssuerByUsername(ctx context.Context, licenseIssuerUsername string) (*model.LicenseIssuer, error) {
	li, err := c.db.SelectLicenseIssuerByUsername(ctx, licenseIssuerUsername)
	return li, handleErrDB(err, "getting license issuer")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetLicenseIssuer(ctx context.Context, licenseIssuerID int) (*model.LicenseIssuer, error) {
	li, err := c.db.SelectLicenseIssuerByID(ctx, licenseIssuerID)
	return li, handleErrDB(err, "getting license issuer by id")
}

// Returns ErrSuperadminImmutable
// Returns ErrInvalidInput
// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) UpdateLicenseIssuer(ctx context.Context, licenseIssuerID int, update map[string]interface{}) error {
	if licenseIssuerID == 0 {
		return ErrSuperadminImmutable
	}
	return c.updateLicenseIssuer(ctx, licenseIssuerID, update)
}

// UpdateLicenseIssuerBypass
//
// Returns ErrInvalidInput
// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) UpdateLicenseIssuerBypass(ctx context.Context, licenseIssuerID int, update map[string]interface{}) error {
	return c.updateLicenseIssuer(ctx, licenseIssuerID, update)
}

func (c *Core) updateLicenseIssuer(ctx context.Context, licenseIssuerID int, update map[string]interface{}) error {
	// Validate username
	v, ok := update["username"]
	if ok {
		username, ok := v.(string)
		if !ok || !ValidUsername(username) {
			return fmt.Errorf("%w username", ErrInvalidInput)
		}
	}
	// Validate max licenses
	v, ok = update["maxLicenses"]
	if ok {
		maxLicenses, ok := v.(model.Limit)
		if !ok {
			return fmt.Errorf("%w max licenses", ErrInvalidInput)
		}
		count, err := c.db.SelectLicensesCountByIssuerID(ctx, licenseIssuerID)
		if err != nil {
			return handleErrDB(err, "counting licenses")
		}
		if !maxLicenses.Allows(count) {
			return fmt.Errorf("%w max licenses: too small", ErrInvalidInput)
		}
	}

	updateApplyRemap(update, licenseIssuerRemap) // Remap json keys to db keys
	update["updated"] = time.Now()
	err := c.db.UpdateLicenseIssuer(ctx, licenseIssuerID, update)
	return handleErrDB(err, "updating license issuer")
}

// Returns ErrSuperadminImmutable
// Returns SensitiveError
func (c *Core) DeleteLicenseIssuer(ctx context.Context, licenseIssuerID int) error {
	if licenseIssuerID == 0 {
		return ErrSuperadminImmutable
	}
	_, err := c.db.DeleteLicenseIssuerByID(ctx, licenseIssuerID)
	return handleErrDB(err, "deleting license issuer")
}

func (c *Core) AuthorizeLicenseIssuerUpdate(login *model.LicenseIssuer) (updateMask []string, delete bool) {
	if c.IsPrivileged(login) {
		// Privileged user can manage most of the account
		return []string{"active", "username", "maxLicenses"}, true
	}
	// Normal user can change only it's username
	return []string{"username"}, false
}
