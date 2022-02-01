package model

import (
	"time"
)

type License struct {
	ID          *[32]byte
	Key         *[32]byte
	Note        string
	CustomData  string
	MaxSessions int
	Created     time.Time
	Updated     time.Time
	LastUsed    *time.Time
	IssuerID    int
}
