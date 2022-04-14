package model

import "time"

type LicenseSession struct {
	ClientID   []byte    `json:"csid"`
	ServerID   []byte    `json:"ssid"`
	ServerKey  []byte    `json:"-"`
	Identifier string    `json:"identifier"`
	MachineID  []byte    `json:"machineID"`
	AppVersion string    `json:"appVersion"`
	Created    time.Time `json:"created"`
	Expire     time.Time `json:"expire"`
	LicenseID  []byte    `json:"-"`
}
