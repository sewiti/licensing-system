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

// Returns ErrInvalidInput
// Returns ErrPasswdTooWeak
// Returns ErrDuplicate
// Returns SensitiveError
func (c *Core) NewLicenseIssuer(ctx context.Context, username, password, email, phoneNumber string, maxLicenses model.Limit) (*model.LicenseIssuer, error) {
	if !ValidUsername(username) {
		return nil, fmt.Errorf("%w username", ErrInvalidInput)
	}
	if email != "" && !ValidEmail(email) {
		return nil, fmt.Errorf("%w email", ErrInvalidInput)
	}
	if phoneNumber != "" && !ValidPhoneNumber(phoneNumber) {
		return nil, fmt.Errorf("%w phone number", ErrInvalidInput)
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
		Email:        email,
		PhoneNumber:  phoneNumber,
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

// // Returns SensitiveError
// func (c *Core) GetPartLicenseIssuers(ctx context.Context, limit, offset int) (lii []*model.LicenseIssuer, total int, err error) {
// 	lii, err = c.db.SelectPartLicenseIssuers(ctx, limit, offset)
// 	if err != nil {
// 		return nil, 0, handleErrDB(err, "getting part of license issuers")
// 	}
// 	total, err = c.db.SelectCountLicenseIssuers(ctx)
// 	return lii, total, handleErrDB(err, "getting count of license issuers")
// }

// // Returns SensitiveError
// func (c *Core) SearchLicenseIssuers(ctx context.Context, username string, limit int) ([]*model.LicenseIssuer, error) {
// 	lii, err := c.db.SelectLicenseIssuersContainUsername(ctx, username, limit)
// 	return lii, handleErrDB(err, "searching license issuers by username")
// }

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
func (c *Core) UpdateLicenseIssuer(ctx context.Context, li *model.LicenseIssuer, changes map[string]struct{}) error {
	if li.ID == 0 {
		return ErrSuperadminImmutable
	}
	return c.updateLicenseIssuer(ctx, li, changes)
}

// UpdateLicenseIssuerBypass
//
// Returns ErrInvalidInput
// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) UpdateLicenseIssuerBypass(ctx context.Context, li *model.LicenseIssuer, changes map[string]struct{}) error {
	return c.updateLicenseIssuer(ctx, li, changes)
}

func (c *Core) updateLicenseIssuer(ctx context.Context, li *model.LicenseIssuer, changes map[string]struct{}) error {
	update := map[string]interface{}{
		"updated": time.Now(),
	}

	if _, ok := changes["active"]; ok {
		update["active"] = li.Active
	}
	if _, ok := changes["username"]; ok {
		if !ValidUsername(li.Username) {
			return fmt.Errorf("%w username", ErrInvalidInput)
		}
		update["username"] = li.Username
	}
	if _, ok := changes["email"]; ok {
		if li.Email != "" && !ValidEmail(li.Email) {
			return fmt.Errorf("%w email", ErrInvalidInput)
		}
		update["email"] = li.Email
	}
	if _, ok := changes["phoneNumber"]; ok {
		if li.PhoneNumber != "" && !ValidPhoneNumber(li.PhoneNumber) {
			return fmt.Errorf("%w phone number", ErrInvalidInput)
		}
		update["phone_number"] = li.PhoneNumber
	}
	if _, ok := changes["maxLicenses"]; ok {
		count, err := c.db.SelectLicensesCountByIssuerID(ctx, li.ID)
		if err != nil {
			return handleErrDB(err, "counting licenses")
		}
		if !li.MaxLicenses.Allows(count) {
			return fmt.Errorf("%w max licenses: too small", ErrInvalidInput)
		}
		update["max_licenses"] = li.MaxLicenses
	}

	err := c.db.UpdateLicenseIssuer(ctx, li.ID, update)
	return handleErrDB(err, "updating license issuer")
}

// Returns ErrSuperadminImmutable
// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) DeleteLicenseIssuer(ctx context.Context, licenseIssuerID int) error {
	if licenseIssuerID == 0 {
		return ErrSuperadminImmutable
	}
	_, err := c.db.DeleteLicenseIssuerByID(ctx, licenseIssuerID)
	return handleErrDB(err, "deleting license issuer")
}

func (c *Core) AuthorizeLicenseIssuerUpdate(login *model.LicenseIssuer) (mask []string, delete bool) {
	if c.IsPrivileged(login) {
		// Privileged user can manage most of the account
		return []string{"active", "username", "email", "phoneNumber", "maxLicenses"}, true
	}
	// Normal user can change only it's contacts
	return []string{"email", "phoneNumber"}, false
}
