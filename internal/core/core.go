// Package core provides main licensing system logic.
package core

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sewiti/licensing-system/internal/core/auth"
	"github.com/sewiti/licensing-system/internal/db"
	"golang.org/x/crypto/nacl/box"
)

type Core struct {
	serverID  *[32]byte
	serverKey *[32]byte

	db  *db.Handler
	lim *limiter
	tm  *auth.TokenManager

	minPasswdEntropy float64

	refresh      RefreshConf
	maxTimeDrift time.Duration
}

type RefreshConf struct {
	Min    time.Duration
	Max    time.Duration
	Jitter float64
}

type LicensingConf struct {
	MaxTimeDrift     time.Duration
	MinPasswdEntropy float64

	Limiter LimiterConf
	Refresh RefreshConf
}

func NewCore(db *db.Handler, serverKey []byte, now time.Time, cfg LicensingConf) (*Core, error) {
	if len(serverKey) != 32 {
		return nil, errors.New("server key must be of length 32")
	}
	if cfg.MinPasswdEntropy < 0 {
		return nil, errors.New("minimum password entropy must be greater or equal to zero")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	id, key, err := box.GenerateKey(bytes.NewBuffer(serverKey))
	if err != nil {
		return nil, err
	}
	tm, err := auth.NewTokenManager(serverKey, hostname)
	if err != nil {
		return nil, err
	}

	return &Core{
		serverID:  id,
		serverKey: key,

		db: db,
		lim: &limiter{
			conf:  cfg.Limiter,
			cache: cache.New(cfg.Limiter.CacheExpiration, cfg.Limiter.CacheCleanupInterval),
		},
		tm: tm,

		minPasswdEntropy: cfg.MinPasswdEntropy,

		refresh:      cfg.Refresh,
		maxTimeDrift: cfg.MaxTimeDrift,
	}, nil
}

func (c *Core) ServerID() *[32]byte {
	id := new([32]byte)
	copy(id[:], c.serverID[:])
	return id
}

func (c *Core) ServerKey() *[32]byte {
	key := new([32]byte)
	copy(key[:], c.serverKey[:])
	return key
}
