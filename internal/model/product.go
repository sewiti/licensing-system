package model

import "time"

type Product struct {
	ID           int       `json:"id"`
	Active       bool      `json:"active"`
	Name         string    `json:"name"`
	ContactEmail string    `json:"contactEmail"`
	Data         []byte    `json:"data"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	IssuerID     int       `json:"-"`
}
