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
