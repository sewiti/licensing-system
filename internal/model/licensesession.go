package model

import "time"

type LicenseSession struct {
	ClientID  *[32]byte
	ServerID  *[32]byte
	ServerKey *[32]byte

	MachineID []byte
	Created   time.Time
	Expire    time.Time
	LicenseID *[32]byte
}
