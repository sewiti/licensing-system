package auth

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vk-rv/pvx"
)

func TestTokenManager_issueToken(t *testing.T) {
	seed := []byte{
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	}
	tm, err := NewTokenManager(seed, "testing.dev")
	require.NoError(t, err)

	type args struct {
		rand     []byte
		subject  string
		now      time.Time
		validFor time.Duration
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "valid token",
			args: args{
				rand:     []byte("labaryta"),
				subject:  "0",
				now:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				validFor: 12 * time.Hour,
			},
			want:      "v4.public.eyJpc3MiOiJ0ZXN0aW5nLmRldiIsInN1YiI6IjAiLCJleHAiOiIyMDIyLTAxLTAxVDEyOjAwOjAwWiIsImlhdCI6IjIwMjItMDEtMDFUMDA6MDA6MDBaIiwianRpIjoiYkdGaVlYSjVkR0Uifeucj9XEIR-dd6w-peSPtpI00t5H-at8h0725pnF2rAe-SJ6cUotUKJOT15nd3AeEOwdOMVuUTIshooRouYhDwI",
			assertion: assert.NoError,
		},
		{
			name: "valid token, indefinite",
			args: args{
				rand:     []byte("labaryta"),
				subject:  "0",
				now:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				validFor: -1,
			},
			want:      "v4.public.eyJpc3MiOiJ0ZXN0aW5nLmRldiIsInN1YiI6IjAiLCJpYXQiOiIyMDIyLTAxLTAxVDAwOjAwOjAwWiIsImp0aSI6ImJHRmlZWEo1ZEdFIn2MUzV-tMsl8lCRLiDgKwzCcnzDQxMY6qY4W70oD3Bh-sTNoXeHG9t0kQIQAPkX-q__0unvOP4oBId0iP6O3dIO",
			assertion: assert.NoError,
		},
		{
			name: "rand too short",
			args: args{
				rand: []byte("laba"),
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand := bytes.NewReader(tt.args.rand)
			got, err := tm.issueToken(rand, tt.args.subject, tt.args.now, tt.args.validFor)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenManager_verifyToken(t *testing.T) {
	seed := []byte{
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
		0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	}
	tm, err := NewTokenManager(seed, "testing.dev")
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		want      *pvx.RegisteredClaims
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:  "expired token",
			token: "v4.public.eyJpc3MiOiJ0ZXN0aW5nLmRldiIsInN1YiI6IjAiLCJleHAiOiIyMDIyLTAxLTAxVDEyOjAwOjAwWiIsImlhdCI6IjIwMjItMDEtMDFUMDA6MDA6MDBaIiwianRpIjoiYkdGaVlYSjVkR0Uifeucj9XEIR-dd6w-peSPtpI00t5H-at8h0725pnF2rAe-SJ6cUotUKJOT15nd3AeEOwdOMVuUTIshooRouYhDwI",
			want: &pvx.RegisteredClaims{
				Issuer:     "testing.dev",
				Subject:    "0",
				IssuedAt:   pvx.TimePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				Expiration: pvx.TimePtr(time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)),
				TokenID:    "bGFiYXJ5dGE", // "labaryta" in base64
			},
			assertion: assert.Error,
		},
		{
			name:  "valid token, indefinite",
			token: "v4.public.eyJpc3MiOiJ0ZXN0aW5nLmRldiIsInN1YiI6IjAiLCJpYXQiOiIyMDIyLTAxLTAxVDAwOjAwOjAwWiIsImp0aSI6ImJHRmlZWEo1ZEdFIn2MUzV-tMsl8lCRLiDgKwzCcnzDQxMY6qY4W70oD3Bh-sTNoXeHG9t0kQIQAPkX-q__0unvOP4oBId0iP6O3dIO",
			want: &pvx.RegisteredClaims{
				Issuer:   "testing.dev",
				Subject:  "0",
				IssuedAt: pvx.TimePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				TokenID:  "bGFiYXJ5dGE", // "labaryta" in base64
			},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.verifyToken(tt.token)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
