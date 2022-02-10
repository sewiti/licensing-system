// Package core provides main licensing system logic.
package core

import (
	"bytes"
	"errors"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sewiti/licensing-system/internal/db"
	"golang.org/x/crypto/nacl/box"
)

type Core struct {
	serverID  *[32]byte
	serverKey *[32]byte

	db  *db.Handler
	lim *limiter

	refresh      RefreshConf
	maxTimeDrift time.Duration

	mathrand *mathrand.Rand
}

type RefreshConf struct {
	Min    time.Duration
	Max    time.Duration
	Jitter float64
}

type LicensingConf struct {
	MaxTimeDrift time.Duration

	Limiter LimiterConf
	Refresh RefreshConf
}

func NewCore(db *db.Handler, serverKey []byte, now time.Time, l LicensingConf) (*Core, error) {
	if len(serverKey) != 32 {
		return nil, errors.New("core: server key must be of length 32")
	}
	id, key, err := box.GenerateKey(bytes.NewBuffer(serverKey))
	if err != nil {
		return nil, fmt.Errorf("core: %w", err)
	}
	return &Core{
		serverID:  id,
		serverKey: key,

		db: db,
		lim: &limiter{
			conf:  l.Limiter,
			cache: cache.New(l.Limiter.CacheExpiration, l.Limiter.CacheCleanupInterval),
		},

		refresh:      l.Refresh,
		maxTimeDrift: l.MaxTimeDrift,

		mathrand: mathrand.New(mathrand.NewSource(now.UnixNano())),
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
