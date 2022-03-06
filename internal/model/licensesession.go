package model

import "time"

type LicenseSession struct {
	ClientID   *[32]byte `json:"csid"`
	ServerID   *[32]byte `json:"ssid"`
	ServerKey  *[32]byte `json:"-"`
	Identifier string    `json:"identifier"`
	MachineID  []byte    `json:"machineID"`
	Created    time.Time `json:"created"`
	Expire     time.Time `json:"expire"`
	LicenseID  *[32]byte `json:"-"`
}
