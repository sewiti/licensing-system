package core

import (
	"context"
	"fmt"
	"time"

	cryptorand "crypto/rand"
	mathrand "math/rand"

	"github.com/sewiti/licensing-system/internal/model"
	"github.com/sewiti/licensing-system/pkg/util"
)

// Returns ErrTimeOutOfSync
// Returns ErrLicenseExpired
// Returns ErrLicenseInactive
// Returns ErrProductInactive
// Returns ErrRateLimitReached
// Returns ErrLicenseIssuerDisabled
// Returns SensitiveError
func (c *Core) NewLicenseSession(ctx context.Context, l *model.License, clientSessionID []byte, identifier string, machineID []byte, appVersion string, clientTime time.Time) (ls *model.LicenseSession, p *model.Product, refresh time.Time, err error) {
	if len(clientSessionID) != 32 {
		return nil, nil, time.Time{}, fmt.Errorf("%w client session id", ErrInvalidInput)
	}
	now := time.Now()
	if !c.timeInSync(now, clientTime) {
		return nil, nil, time.Time{}, ErrTimeOutOfSync
	}
	if !l.Active {
		return nil, nil, time.Time{}, ErrLicenseInactive
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(now) {
		return nil, nil, time.Time{}, ErrLicenseExpired
	}
	rl := c.lim.get(l)
	if !rl.Allow() {
		return nil, nil, time.Time{}, ErrRateLimitReached
	}
	li, err := c.GetLicenseIssuer(ctx, l.IssuerID)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if !li.Active {
		return nil, nil, time.Time{}, ErrLicenseIssuerDisabled
	}
	if l.ProductID != nil {
		p, err = c.GetProduct(ctx, *l.ProductID)
		if err != nil {
			return nil, nil, time.Time{}, err
		}
		if !p.Active {
			return nil, nil, time.Time{}, ErrProductInactive
		}
	} else {
		p = &model.Product{}
	}
	// Max sessions are taken care of by the cleanup routine.

	serverID, serverKey, err := util.GenerateKey(cryptorand.Reader)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	refresh, expiry := c.calcLicenseSessionTimes(now, now)
	s := &model.LicenseSession{
		ClientID:   clientSessionID,
		ServerID:   serverID,
		ServerKey:  serverKey,
		Identifier: identifier,
		MachineID:  machineID,
		AppVersion: appVersion,
		Created:    now,
		Expire:     expiry,
		LicenseID:  l.ID,
	}
	// // Delete old client's license sessions
	// _, err = c.db.DeleteLicenseSessionsByLicenseIDAndMachineID(ctx, l.ID, machineID)
	// err = handleErrDB(err, "deleting old license sessions")
	// if err != nil && !errors.Is(err, ErrNotFound) {
	// 	return nil, time.Time{}, err
	// }
	err = c.db.UpdateLicense(ctx, l.ID, l.IssuerID, map[string]interface{}{
		"last_used": now,
	})
	err = handleErrDB(err, "updating license")
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	err = c.db.InsertLicenseSession(ctx, s)
	return s, p, refresh, handleErrDB(err, "creating license session")
}

// Returns SensitiveError
func (c *Core) GetAllLicenseSessionsByLicense(ctx context.Context, licenseID []byte) ([]*model.LicenseSession, error) {
	lss, err := c.db.SelectAllLicenseSessionsByLicenseID(ctx, licenseID)
	return lss, handleErrDB(err, "getting all license sessions")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) GetLicenseSession(ctx context.Context, clientSessionID []byte) (*model.LicenseSession, error) {
	ls, err := c.db.SelectLicenseSessionByID(ctx, clientSessionID)
	return ls, handleErrDB(err, "getting license session")
}

// Returns ErrTimeOutOfSync
// Returns ErrLicenseExpired
// Returns ErrLicenseInactive
// Returns ErrProductInactive
// Returns ErrLicenseIssuerDisabled
// Returns ErrLicenseSessionExpired
// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) UpdateLicenseSession(ctx context.Context, ls *model.LicenseSession, l *model.License, clientTime time.Time) (p *model.Product, refresh time.Time, err error) {
	now := time.Now()
	if !c.timeInSync(now, clientTime) {
		return nil, time.Time{}, ErrTimeOutOfSync
	}
	if !l.Active {
		return nil, time.Time{}, ErrLicenseInactive
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(now) {
		return nil, time.Time{}, ErrLicenseExpired
	}
	if now.After(ls.Expire) {
		return nil, time.Time{}, ErrLicenseSessionExpired
	}
	li, err := c.GetLicenseIssuer(ctx, l.IssuerID)
	if err != nil {
		return nil, time.Time{}, err
	}
	if !li.Active {
		return nil, time.Time{}, ErrLicenseIssuerDisabled
	}
	if l.ProductID != nil {
		p, err = c.GetProduct(ctx, *l.ProductID)
		if err != nil {
			return nil, time.Time{}, err
		}
		if !p.Active {
			return nil, time.Time{}, ErrProductInactive
		}
	} else {
		p = &model.Product{}
	}
	// Max sessions are taken care of by the cleanup routine.

	refresh, expiry := c.calcLicenseSessionTimes(ls.Created, now)
	ls.Expire = expiry

	err = c.db.UpdateLicenseSession(ctx, ls)
	return p, refresh, handleErrDB(err, "updating license session")
}

// Returns ErrNotFound
// Returns SensitiveError
func (c *Core) DeleteLicenseSession(ctx context.Context, clientSessionID []byte) error {
	// We don't care about client time when deleting session.
	_, err := c.db.DeleteLicenseSessionBySessionID(ctx, clientSessionID)
	return handleErrDB(err, "deleting license session")
}

// timeInSync reports whether client time is in sync with server time, i. e,
// haven't drifted from server time too far (defined by c.maxTimeDrift).
func (c *Core) timeInSync(server, client time.Time) bool {
	lowerBound := server.Add(-c.maxTimeDrift)
	upperBound := server.Add(c.maxTimeDrift)
	return client.After(lowerBound) && client.Before(upperBound)
}

// calcLicenseSessionTimes calculates license session refresh and expire times.
//
//  Refresh time = 2 * uptime (+-jitter%, clamped to min-max)
//  Expire time  = 2 * refresh time
func (c *Core) calcLicenseSessionTimes(start, now time.Time) (refresh, expiry time.Time) {
	// Random [-jitter; +jitter)
	jitter := (2.0 * c.refresh.Jitter * mathrand.Float64()) - c.refresh.Jitter

	uptime := now.Sub(start)
	delay := time.Duration(
		(2.0 + jitter) * float64(uptime), // 2.0 * uptime
	)

	// Clamp to [min; max]
	if delay < c.refresh.Min {
		delay = c.refresh.Min
	} else if delay > c.refresh.Max {
		delay = c.refresh.Max
	}
	return now.Add(delay), now.Add(2 * delay)
}
