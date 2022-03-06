package model

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit_MarshalJSON(t *testing.T) {
	tests := []struct {
		arg  Limit
		want string
	}{
		{
			arg:  1,
			want: "1",
		},
		{
			arg:  0,
			want: "null",
		},
		{
			arg:  -1,
			want: "null",
		},
	}
	for _, tt := range tests {
		t.Run(strconv.Itoa(int(tt.arg)), func(t *testing.T) {
			got, err := tt.arg.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestLimit_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		arg       string
		want      Limit
		assertion assert.ErrorAssertionFunc
	}{
		{
			arg:       "0",
			want:      0,
			assertion: assert.NoError,
		},
		{
			arg:       "null",
			want:      0,
			assertion: assert.NoError,
		},
		{
			arg:       "-1",
			want:      0,
			assertion: assert.NoError,
		},
		{
			arg:       "1",
			want:      1,
			assertion: assert.NoError,
		},
		{
			arg:       "abc",
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			var l Limit
			err := l.UnmarshalJSON([]byte(tt.arg))
			tt.assertion(t, err)
		})
	}
}

func TestLimitJSON(t *testing.T) {
	tests := []string{
		"null",
		"1",
		"12",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			var l Limit
			err := json.Unmarshal([]byte(tt), &l)
			assert.NoError(t, err)

			bs, err := json.Marshal(l)
			assert.NoError(t, err)
			assert.Equal(t, tt, string(bs))
		})
	}
}
