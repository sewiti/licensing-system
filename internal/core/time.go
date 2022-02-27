package core

import (
	"time"
)

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
	jitter := (2.0 * c.refresh.Jitter * c.mathrand.Float64()) - c.refresh.Jitter

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
