package core

import (
	"context"

	"github.com/sewiti/licensing-system/pkg/model"
)

func (c *Core) GetLicense(ctx context.Context, licenseID *[32]byte) (*model.License, error) {
	l, err := c.db.SelectLicenseByID(ctx, licenseID)
	if err != nil {
		return nil, &SensitiveError{
			Message: "getting license",
			err:     err,
		}
	}
	return l, nil
}
