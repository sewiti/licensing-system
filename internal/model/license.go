package model

import (
	"encoding/json"
	"time"
)

type License struct {
	ID  *[32]byte
	Key *[32]byte

	Note string

	Data        json.RawMessage
	MaxSessions int
	ValidUntil  *time.Time

	Created time.Time
	Updated time.Time

	IssuerID int
}
