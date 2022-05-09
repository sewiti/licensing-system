package core

import (
	"testing"

	"github.com/sewiti/licensing-system/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCore_IsPrivileged(t *testing.T) {
	tests := []struct {
		name string
		li   *model.LicenseIssuer
		want bool
	}{
		{
			name: "nil",
			li:   nil,
			want: false,
		},
		{
			name: "privileged",
			li: &model.LicenseIssuer{
				ID: 0,
			},
			want: true,
		},
		{
			name: "not privileged",
			li: &model.LicenseIssuer{
				ID: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&Core{}).IsPrivileged(tt.li)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCore_SufficientPasswdStrength(t *testing.T) {
	c := &Core{
		minPasswdEntropy: 30,
	}
	tests := []struct {
		username    string
		password    string
		wantEntropy float64
		wantOk      bool
	}{
		{
			username:    "boi",
			password:    "boi",
			wantEntropy: 0.0,
			wantOk:      false,
		},
		{
			username:    "user",
			password:    "helloHowAreYou",
			wantEntropy: 19.832,
			wantOk:      false,
		},
		{
			username:    "strong-user",
			password:    "this-is-a-very-long-password-which-is-good-enough",
			wantEntropy: 101.628,
			wantOk:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			gotEntropy, gotOk := c.SufficientPasswdStrength(tt.username, tt.password)
			assert.Equal(t, tt.wantEntropy, gotEntropy)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
