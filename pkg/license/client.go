package license

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/nacl/box"
)

const DefaultMaxRefresh = 24 * time.Hour

type Client struct {
	licenseID  *[32]byte
	licenseKey *[32]byte
	serverID   *[32]byte // Server ID, not to be confused with Server Session ID.
	machineID  []byte
	url        string

	state State

	mu      sync.RWMutex
	session *session

	rand io.Reader
}

var ErrNotConnected = errors.New("license: client: session not established")

func NewClient(url string, serverID, machineID, licenseKey []byte) (*Client, error) {
	if len(serverID) != 32 {
		return nil, errors.New("license: client: server id must be of length 32")
	}
	serverIDCopy := new([32]byte)
	copy(serverIDCopy[:], serverID)

	if len(licenseKey) != 32 {
		return nil, errors.New("license: client: license key must be of length 32")
	}
	clientID, clientKey, err := box.GenerateKey(bytes.NewReader(licenseKey))
	if err != nil {
		return nil, fmt.Errorf("license: client: generating session keys: %w", err)
	}

	return &Client{
		licenseID:  clientID,
		licenseKey: clientKey,
		serverID:   serverIDCopy,
		machineID:  machineID,
		url:        url,

		state: StateInvalid,

		rand: cryptorand.Reader,
	}, nil
}

func NewClientFiles(url string, serverID []byte, machineIDFile, licenseFile string) (*Client, error) {
	machineID, err := ReadID(machineIDFile)
	if err != nil {
		return nil, fmt.Errorf("license: client: %w", err)
	}
	licenseKey, err := ReadKey(licenseFile)
	if err != nil {
		return nil, fmt.Errorf("license: client: %w", err)
	}
	return NewClient(url, serverID, machineID, licenseKey)
}

func (c *Client) newSession(ctx context.Context) (*session, error) {
	clientID, clientKey, err := box.GenerateKey(c.rand)
	if err != nil {
		return nil, fmt.Errorf("license: session-create: %w", err)
	}
	data, err := c.sendCreateSession(ctx, clientID, clientKey, c.rand)
	if err != nil {
		return nil, fmt.Errorf("license: session-create: %w", err)
	}
	s := &session{
		serverID:  data.ServerSessionID,
		clientID:  clientID,
		clientKey: clientKey,
		url:       c.url,
	}
	s.updateTimes(time.Now(), data.Timestamp, data.RefreshAfter, data.ExpireAfter)
	return s, nil
}

func (c *Client) Run(ctx context.Context, maxRefresh time.Duration, cb SessionCallback) {
	const (
		retryIn    = 30 * time.Second
		retryInMax = 30 * time.Minute
	)
	if maxRefresh <= 0 {
		panic(fmt.Errorf("maxRefresh must be greater than zero: %v", maxRefresh))
	}
	if ctx.Err() != nil {
		return
	}

	// Repeatedly try to create license session
	retryDelay := retryIn
	for {
		c.mu.Lock()
		s, err := c.newSession(ctx)
		if err == nil {
			c.session = s
			c.state = StateValid
			c.mu.Unlock()
			retryDelay = retryIn
			break
		}
		c.state = StateInvalid

		if _, ok := err.(temporaryError); !ok {
			// Error
			cb.call("creating license session", err)
			c.mu.Unlock()
			return
		}
		// Temporary error - schedule a retry
		cb.call(fmt.Sprintf("creating license session, retrying in %v", retryDelay), err)
		c.mu.Unlock()

		select {
		case <-time.After(retryDelay):
			retryDelay *= 2
			if retryDelay > retryInMax {
				retryDelay = retryInMax
			}
		case <-ctx.Done():
			return
		}
	}

	// Repeatedly refresh license session
	for {
		now := time.Now()
		refreshAfter := now.Sub(c.session.refreshAfter)
		if refreshAfter > maxRefresh {
			refreshAfter = maxRefresh
		}
		refreshT := time.NewTimer(refreshAfter)
		expireT := time.NewTimer(now.Sub(c.session.expireAfter))

		select {
		case <-refreshT.C: // Needs refreshing
			expireT.Stop()

			c.mu.Lock()
			err := c.session.refresh(ctx, c.rand)
			if err == nil {
				c.state = StateValid
				cb.call("license session refreshed successfully", nil)
				c.mu.Unlock()
				retryDelay = retryIn // Reset delay
				continue
			}

			if _, ok := err.(temporaryError); !ok {
				// Error
				cb.call("refreshing license session", err)
				c.mu.Unlock()
				return
			}
			// Temporary error - schedule a retry
			c.session.refreshAfter = time.Now().Add(retryDelay)
			// No changes to license session state
			cb.call(fmt.Sprintf("refreshing license session, retrying in %v", retryDelay), err)
			c.mu.Unlock()

			retryDelay *= 2
			if retryDelay > retryInMax {
				retryDelay = retryInMax
			}

		case <-expireT.C: // Expired
			refreshT.Stop()

			c.mu.Lock()
			// No need for sending any request
			c.state = StateExpired
			cb.call("license session has expired", nil)
			c.mu.Unlock()
			return

		case <-ctx.Done(): // App closed
			refreshT.Stop()
			expireT.Stop()

			c.mu.Lock()
			err := c.session.close(ctx, c.rand)
			c.session = nil // Get rid of session
			c.state = StateClosed
			if err != nil {
				cb.call("closing license session", err)
				// Continue on error
			} else {
				cb.call("license session closed successfully", nil)
			}
			c.mu.Unlock()
			return
		}
	}
}

func (c *Client) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Client) UnmarshalData(v interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.session == nil {
		return ErrNotConnected
	}
	return json.Unmarshal(c.session.data, v)
}

type SessionCallback func(msg string, err error)

func (cb SessionCallback) call(msg string, err error) {
	if cb != nil {
		cb(msg, err)
	}
}
