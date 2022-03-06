package model

import "time"

type LicenseIssuer struct {
	ID           int       `json:"id"`
	Active       bool      `json:"active"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	MaxLicenses  Limit     `json:"maxLicenses"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}
