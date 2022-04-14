package license

import (
	"context"
	"fmt"
	"io"
	"time"
)

type session struct {
	serverID  []byte // Server Session ID
	clientID  []byte // Client Session ID
	clientKey []byte // Client Session Key
	url       string

	refreshAfter time.Time
	expireAfter  time.Time

	data []byte
}

func (s *session) updateTimes(now, remote, refreshAfter, expireAfter time.Time) {
	s.refreshAfter = now.Add(refreshAfter.Sub(remote))
	s.expireAfter = now.Add(expireAfter.Sub(remote))
}

func (s *session) refresh(ctx context.Context, rand io.Reader) error {
	data, err := s.sendRefresh(ctx, rand)
	if err != nil {
		return fmt.Errorf("license: session-refresh: %w", err)
	}
	s.updateTimes(time.Now(), data.Timestamp, data.RefreshAfter, data.ExpireAfter)
	return nil
}

func (s *session) close(ctx context.Context, rand io.Reader) error {
	err := s.sendClose(ctx, rand)
	if err != nil {
		return fmt.Errorf("license: session-close: %w", err)
	}
	return nil
}
