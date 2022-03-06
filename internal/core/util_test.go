package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateInMask(t *testing.T) {
	tests := []struct {
		name         string
		update       map[string]interface{}
		mask         []string
		wantBadField string
		wantOk       bool
	}{
		{
			name: "ok",
			update: map[string]interface{}{
				"username": "hello",
			},
			mask:   []string{"username"},
			wantOk: true,
		},
		{
			name: "field not in mask",
			update: map[string]interface{}{
				"active":   true,
				"username": "hello",
			},
			mask:         []string{"username"},
			wantBadField: "active",
			wantOk:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBadField, gotOk := UpdateInMask(tt.update, tt.mask)
			assert.Equal(t, tt.wantBadField, gotBadField)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func Test_updateApplyRemap(t *testing.T) {
	tests := []struct {
		name   string
		update map[string]interface{}
		remap  map[string]string
		want   map[string]interface{}
	}{
		{
			name: "ok",
			update: map[string]interface{}{
				"maxLicenses": 4,
				"username":    "hello",
			},
			remap: licenseIssuerRemap,
			want: map[string]interface{}{
				"max_licenses": 4,
				"username":     "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateApplyRemap(tt.update, tt.remap)
			assert.Equal(t, tt.want, tt.update)
		})
	}
}
