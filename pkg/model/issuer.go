package model

import "time"

type LicenseIssuer struct {
	ID           int
	Active       bool
	Username     string
	PasswordHash string
	MaxLicenses  int
	Created      time.Time
	LastActive   time.Time
}
