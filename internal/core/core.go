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
	"github.com/sewiti/licensing-system/pkg/util"
)

type Core struct {
	serverID  []byte
	serverKey []byte

	db  *db.Handler
	lim *limiter
	tm  *auth.TokenManager

	minPasswdEntropy float64
	useGui           bool

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
	UseGUI           bool

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
	id, key, err := util.GenerateKey(bytes.NewBuffer(serverKey))
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
		useGui:           cfg.UseGUI,

		refresh:      cfg.Refresh,
		maxTimeDrift: cfg.MaxTimeDrift,
	}, nil
}

func (c *Core) ServerID() []byte {
	id := make([]byte, 32)
	copy(id, c.serverID[:])
	return id
}

func (c *Core) ServerKey() []byte {
	key := make([]byte, 32)
	copy(key, c.serverKey[:])
	return key
}

func (c *Core) UseGUI() bool {
	return c.useGui
}
