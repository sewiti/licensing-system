package license

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	cryptorand "crypto/rand"

	"github.com/sewiti/licensing-system/pkg/util"
)

const DefaultMaxRefresh = 24 * time.Hour

type Client struct {
	licenseID  []byte
	licenseKey []byte
	serverID   []byte // Server ID, not to be confused with Server Session ID.
	identifier string
	machineID  []byte
	appVersion string
	url        string

	state State

	mx      sync.RWMutex
	session *session
}

var ErrNotConnected = errors.New("license: client: session not established")

func NewClient(url string, serverID, machineID, licenseKey []byte) (*Client, error) {
	if len(serverID) != 32 {
		return nil, errors.New("license: client: server id must be of length 32")
	}
	identifier, _ := Identifier()
	serverIDCopy := make([]byte, 32)
	copy(serverIDCopy, serverID)

	if len(licenseKey) != 32 {
		return nil, errors.New("license: client: license key must be of length 32")
	}
	clientID, clientKey, err := util.GenerateKey(bytes.NewReader(licenseKey))
	if err != nil {
		return nil, fmt.Errorf("license: client: generating session keys: %w", err)
	}
	return &Client{
		licenseID:  clientID,
		licenseKey: clientKey,
		serverID:   serverIDCopy,
		identifier: identifier,
		machineID:  machineID,
		url:        url,

		state: StateInvalid,
	}, nil
}

func (c *Client) SetIdentifier(id string) {
	c.identifier = id
}

func (c *Client) SetAppVersion(appVersion string) {
	c.appVersion = appVersion
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
	clientID, clientKey, err := util.GenerateKey(cryptorand.Reader)
	if err != nil {
		return nil, fmt.Errorf("license: session-create: %w", err)
	}
	data, err := c.sendCreateSession(ctx, clientID, clientKey, cryptorand.Reader)
	if err != nil {
		return nil, fmt.Errorf("license: session-create: %w", err)
	}
	s := &session{
		serverID:  data.ServerSessionID,
		clientID:  clientID,
		clientKey: clientKey,
		url:       c.url,

		productName: data.ProductName,
		productData: data.ProductData,
		data:        data.Data,
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
		c.mx.Lock()
		s, err := c.newSession(ctx)
		if err == nil {
			c.session = s
			c.state = StateValid
			c.mx.Unlock()
			cb.call("created license session", nil)
			retryDelay = retryIn
			break
		}
		c.state = StateInvalid

		if !errors.Is(err, errTemporary) {
			// Error
			c.mx.Unlock()
			cb.call("creating license session", err)
			return
		}
		// Temporary error - schedule a retry
		c.mx.Unlock()
		cb.call(fmt.Sprintf("creating license session, retrying in %v", retryDelay), err)

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
		refreshAfter := c.session.refreshAfter.Sub(now)
		if refreshAfter > maxRefresh {
			refreshAfter = maxRefresh
		}
		expireAfter := c.session.expireAfter.Sub(now)

		refreshT := time.NewTimer(refreshAfter)
		expireT := time.NewTimer(expireAfter)

		select {
		case <-refreshT.C: // Needs refreshing
			expireT.Stop()

			c.mx.Lock()
			err := c.session.refresh(ctx, cryptorand.Reader)
			if err == nil {
				c.state = StateValid
				c.mx.Unlock()
				cb.call("license session refreshed successfully", nil)
				retryDelay = retryIn // Reset delay
				continue
			}

			if !errors.Is(err, errTemporary) {
				// Error
				_ = c.session.close(ctx, cryptorand.Reader)
				c.session = nil
				c.state = StateClosed
				c.mx.Unlock()
				cb.call("refreshing license session", err)
				return
			}
			// Temporary error - schedule a retry
			c.session.refreshAfter = time.Now().Add(retryDelay)
			// No changes to license session state
			if retryDelay > maxRefresh {
				retryDelay = maxRefresh
			}
			c.mx.Unlock()
			cb.call(fmt.Sprintf("refreshing license session, retrying in %v", retryDelay), err)

			retryDelay *= 2
			if retryDelay > retryInMax {
				retryDelay = retryInMax
			}

		case <-expireT.C: // Expired
			refreshT.Stop()

			c.mx.Lock()
			_ = c.session.close(ctx, cryptorand.Reader)
			c.state = StateExpired
			c.mx.Unlock()
			cb.call("license session has expired", nil)
			return

		case <-ctx.Done(): // App closed
			refreshT.Stop()
			expireT.Stop()

			c.mx.Lock()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := c.session.close(ctx, cryptorand.Reader)
			c.session = nil // Get rid of session
			c.state = StateClosed
			c.mx.Unlock()
			if err != nil {
				cb.call("closing license session", err)
				// Continue on error
			} else {
				cb.call("license session closed successfully", nil)
			}
			return
		}
	}
}

func (c *Client) State() State {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.state
}

func (c *Client) Data() ([]byte, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.session == nil {
		return nil, ErrNotConnected
	}
	data := make([]byte, len(c.session.data))
	copy(data, c.session.data)
	return data, nil
}

func (c *Client) UnmarshalData(v interface{}) error {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.session == nil {
		return ErrNotConnected
	}
	return json.Unmarshal(c.session.data, v)
}

func (c *Client) ProductName() (string, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.session == nil {
		return "", ErrNotConnected
	}
	return c.session.productName, nil
}

func (c *Client) ProductData() ([]byte, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.session == nil {
		return nil, ErrNotConnected
	}
	data := make([]byte, len(c.session.productData))
	copy(data, c.session.productData)
	return data, nil
}

func (c *Client) UnmarshalProductData(v interface{}) error {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.session == nil {
		return ErrNotConnected
	}
	return json.Unmarshal(c.session.productData, v)
}

type SessionCallback func(msg string, err error)

func (cb SessionCallback) call(msg string, err error) {
	if cb != nil {
		cb(msg, err)
	}
}
