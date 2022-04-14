package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangesInMask(t *testing.T) {
	tests := []struct {
		name         string
		changes      map[string]struct{}
		mask         []string
		wantBadField string
		wantOk       bool
	}{
		{
			name: "ok",
			changes: map[string]struct{}{
				"username": {},
			},
			mask:   []string{"username"},
			wantOk: true,
		},
		{
			name: "field not in mask",
			changes: map[string]struct{}{
				"active":   {},
				"username": {},
			},
			mask:         []string{"username"},
			wantBadField: "active",
			wantOk:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBadField, gotOk := ChangesInMask(tt.changes, tt.mask)
			assert.Equal(t, tt.wantBadField, gotBadField)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
