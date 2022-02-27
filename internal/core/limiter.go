package core

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sewiti/licensing-system/internal/model"
	"golang.org/x/time/rate"
)

type LimiterConf struct {
	SessionEvery     time.Duration
	SessionEveryInit time.Duration
	BurstTotal       time.Duration

	CacheExpiration      time.Duration
	CacheCleanupInterval time.Duration
}

type limiter struct {
	conf LimiterConf

	mu    sync.RWMutex
	cache *cache.Cache
}

func (lim *limiter) get(l *model.License) *rate.Limiter {
	id := fmt.Sprintf("%s:%d",
		base64.StdEncoding.EncodeToString(l.ID[:]), l.MaxSessions)
	lim.mu.RLock()
	v, exists := lim.cache.Get(id)
	if exists {
		lim.mu.RUnlock()
		return v.(*rate.Limiter)
	}
	lim.mu.RUnlock()

	lim.mu.Lock()
	rl := lim.newRateLimiter(l.MaxSessions)
	lim.cache.Set(id, rl, cache.DefaultExpiration)
	lim.mu.Unlock()
	return rl
}

func (lim *limiter) newRateLimiter(maxSessions int) *rate.Limiter {
	// For multi-session licenses, proportionally increase allowed session
	// frequency.
	sessionEvery := lim.conf.SessionEvery / time.Duration(maxSessions)

	// Allow bursts of BurstTotal worth of sessions time.
	burst := int(lim.conf.BurstTotal / sessionEvery)
	rl := rate.NewLimiter(rate.Every(sessionEvery), burst)

	// Make initial burst a minimum to support new session every SessionEveryInit.
	burstInit := int(sessionEvery / lim.conf.SessionEveryInit)
	if burstInit < burst {
		rl.AllowN(time.Now(), burst-burstInit) // burst - (burst - burstInit) = burstInit
	}
	return rl
}
