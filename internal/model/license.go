package model

import (
	"encoding/json"
	"time"
)

type License struct {
	ID          *[32]byte       `json:"id"`
	Key         *[32]byte       `json:"key"`
	Note        string          `json:"note"`
	Data        json.RawMessage `json:"data"`
	MaxSessions int             `json:"maxSessions"`
	ValidUntil  *time.Time      `json:"validUntil"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
	IssuerID    int             `json:"-"`
}
