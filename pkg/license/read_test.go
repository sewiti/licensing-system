package license

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadID(t *testing.T) {
	tests := []struct {
		path      string
		want      []byte
		assertion assert.ErrorAssertionFunc
	}{
		{
			path: "testdata/machine-id",
			want: []byte{
				0x9, 0xb7, 0xed, 0xfc, 0x96, 0x75, 0x4b, 0x2c,
				0xb1, 0x32, 0x2, 0x97, 0x94, 0xf3, 0x4d, 0x13,
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := ReadID(tt.path)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadKey(t *testing.T) {
	tests := []struct {
		path      string
		want      []byte
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := ReadKey(tt.path)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
