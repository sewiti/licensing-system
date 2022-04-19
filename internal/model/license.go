package model

import (
	"time"
)

type License struct {
	ID           []byte     `json:"id"`
	Key          []byte     `json:"key"`
	Active       bool       `json:"active"`
	Name         string     `json:"name"`
	Tags         []string   `json:"tags"`
	EndUserEmail string     `json:"endUserEmail"`
	Note         string     `json:"note"`
	Data         []byte     `json:"data"`
	MaxSessions  int        `json:"maxSessions"`
	ValidUntil   *time.Time `json:"validUntil"`
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	LastUsed     *time.Time `json:"lastUsed"`
	IssuerID     int        `json:"-"`
}
