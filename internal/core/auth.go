package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/nbutton23/zxcvbn-go"
	"github.com/sewiti/licensing-system/internal/core/auth"
	"github.com/sewiti/licensing-system/internal/model"
)

// Returns ErrNotFound
// Returns ErrUserInactive
// Returns auth.ErrInvalidPassword
// Returns SensitiveError
func (c *Core) AuthenticateBasic(ctx context.Context, username, passwd string) (*model.LicenseIssuer, error) {
	li, err := c.GetLicenseIssuerByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if !li.Active {
		return nil, ErrUserInactive
	}

	err = auth.VerifyPasswd(passwd, li.PasswordHash)
	if err != nil {
		return nil, err
	}
	return li, nil
}

// Returns ErrNotFound
// Returns ErrUserInactive
// Returns ErrUserInactive
// Returns ErrInvalidToken
// Returns SensitiveError
func (c *Core) AuthenticateToken(ctx context.Context, token string) (*model.LicenseIssuer, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	subject, err := c.tm.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	issuerID, err := strconv.Atoi(subject)
	if err != nil {
		return nil, err
	}

	li, err := c.GetLicenseIssuer(ctx, issuerID)
	if err != nil {
		return nil, err
	}
	if !li.Active {
		return nil, ErrUserInactive
	}
	return li, nil
}

// CreateToken
//
// Returns SensitiveError
func (c *Core) CreateToken(li *model.LicenseIssuer) (string, error) {
	token, err := c.tm.IssueToken(strconv.Itoa(li.ID))
	if err != nil {
		return "", &SensitiveError{Err: err, Message: "creating token"}
	}
	return token, nil
}

// ChangePasswd
//
// Returns ErrNotFound
// Returns ErrInvalidPasswd (old password)
// Returns ErrPasswdTooWeak (new password)
// Returns SensitiveError
func (c *Core) ChangePasswd(ctx context.Context, login *model.LicenseIssuer, licenseIssuerID int, oldPasswd, newPasswd string) error {
	li, err := c.GetLicenseIssuer(ctx, licenseIssuerID)
	if err != nil {
		return handleErrDB(err, "getting license issuer")
	}
	entropy, ok := c.SufficientPasswdStrength(li.Username, newPasswd)
	if !ok {
		return fmt.Errorf("%w: entropy %.2f", ErrPasswdTooWeak, entropy)
	}
	if !c.IsPrivileged(login) {
		err = auth.VerifyPasswd(oldPasswd, li.PasswordHash)
		if err != nil {
			if !errors.Is(err, auth.ErrNoLogin) {
				return err
			}
		}
	}

	passwdHash, err := auth.HashPasswd(newPasswd)
	if err != nil {
		return err
	}
	err = c.db.UpdateLicenseIssuer(ctx, licenseIssuerID, map[string]interface{}{
		"password_hash": passwdHash,
	})
	return handleErrDB(err, "updating license issuer")
}

func (c *Core) SufficientPasswdStrength(username, password string) (entropy float64, ok bool) {
	// userInputs is an additional dictionary for password strength estimation
	// which is used to penalize passwords containing this information.
	userInputs := []string{
		username,
		"online",
		"software",
		"licence",
		"license",
		"licensing",
		"system",
		"session",
		"server",
		"DRM",
		"digital",
		"rights",
		"management",
		"admin",
	}
	str := zxcvbn.PasswordStrength(password, userInputs)
	return str.Entropy, str.Entropy >= c.minPasswdEntropy
}

func (c *Core) IsPrivileged(li *model.LicenseIssuer) bool {
	if li == nil {
		return false
	}
	//  0 - superadmin
	// -1 - cli
	return li.ID == 0 || li.ID == -1
}
