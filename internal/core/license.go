package core

import (
	"context"
	"errors"

	"github.com/sewiti/licensing-system/internal/db"
	"github.com/sewiti/licensing-system/internal/model"
)

func (c *Core) GetLicense(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	l, err := c.db.SelectLicenseByID(ctx, licenseID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, &SensitiveError{
			Message: "getting license",
			err:     err,
		}
	}
	return l, nil
}
