package core

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"time"

	"github.com/sewiti/licensing-system/internal/db"
	"github.com/sewiti/licensing-system/internal/model"
	"golang.org/x/crypto/nacl/box"
)

func (c *Core) NewLicenseSession(ctx context.Context, l *model.License, clientSessionID *[32]byte, machineID []byte, clientTime time.Time) (ls *model.LicenseSession, refresh time.Time, err error) {
	now := time.Now()
	if !c.timeInSync(now, clientTime) {
		return nil, time.Time{}, ErrTimeOutOfSync
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(now) {
		return nil, time.Time{}, ErrLicenseExpired
	}
	rl := c.lim.get(l)
	if !rl.Allow() {
		return nil, time.Time{}, ErrRateLimitReached
	}
	// Max sessions are taken care of by the cleanup routine.

	serverID, serverKey, err := box.GenerateKey(cryptorand.Reader)
	if err != nil {
		return nil, time.Time{}, err
	}

	refresh, expiry := c.calcLicenseSessionTimes(now, now)
	s := &model.LicenseSession{
		ClientID:  clientSessionID,
		ServerID:  serverID,
		ServerKey: serverKey,
		MachineID: machineID,
		Created:   now,
		Expire:    expiry,
		LicenseID: l.ID,
	}
	// Delete old client's license sessions
	_, err = c.db.DeleteLicenseSessionsByLicenseIDAndMachineID(ctx, l.ID, machineID)
	if err != nil {
		return nil, time.Time{}, &SensitiveError{
			Message: "creating license session",
			err:     err,
		}
	}
	err = c.db.InsertLicenseSession(ctx, s)
	if err != nil {
		return nil, time.Time{}, &SensitiveError{
			Message: "creating license session",
			err:     err,
		}
	}
	return s, refresh, nil
}

func (c *Core) GetLicenseSession(ctx context.Context, clientSessionID *[32]byte) (*model.LicenseSession, error) {
	ls, err := c.db.SelectLicenseSessionByID(ctx, clientSessionID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, &SensitiveError{
			Message: "retrieving license session",
			err:     err,
		}
	}
	return ls, nil
}

func (c *Core) GetLicenseSessionsCount(ctx context.Context, licenseID *[32]byte) (int, error) {
	count, err := c.db.SelectLicenseSessionsCountByLicenseID(ctx, licenseID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, &SensitiveError{
			Message: "retrieving license sessions count",
			err:     err,
		}
	}
	return count, nil
}

func (c *Core) UpdateLicenseSession(ctx context.Context, ls *model.LicenseSession, l *model.License, clientTime time.Time) (refresh time.Time, err error) {
	now := time.Now()
	if !c.timeInSync(now, clientTime) {
		return time.Time{}, ErrTimeOutOfSync
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(now) {
		return time.Time{}, ErrLicenseExpired
	}
	if now.After(ls.Expire) {
		return time.Time{}, ErrLicenseSessionExpired
	}
	// Max sessions are taken care of by the cleanup routine.

	refresh, expiry := c.calcLicenseSessionTimes(ls.Created, now)
	ls.Expire = expiry

	err = c.db.UpdateLicenseSession(ctx, ls)
	if err != nil {
		return time.Time{}, &SensitiveError{
			Message: "updating license session",
			err:     err,
		}
	}
	return refresh, nil
}

func (c *Core) DeleteLicenseSession(ctx context.Context, clientSessionID *[32]byte) error {
	// We don't care about client time when deleting session.
	_, err := c.db.DeleteLicenseSessionBySessionID(ctx, clientSessionID)
	if err != nil {
		return &SensitiveError{
			Message: "deleting license session",
			err:     err,
		}
	}
	return nil
}
