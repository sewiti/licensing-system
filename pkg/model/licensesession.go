package model

import "time"

type LicenseSession struct {
	ID          *[32]byte
	Key         *[32]byte
	MachineUUID []byte
	Created     time.Time
	Expire      time.Time
	LicenseID   *[32]byte
}
