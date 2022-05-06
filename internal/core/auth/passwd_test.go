package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_hashPasswd(t *testing.T) {
	tests := []struct {
		password string
		salt     []byte
		want     string
	}{
		{
			password: "qwerty",
			salt: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			want: "argon2id$v=19,t=4,m=65536,p=4$AAAAAAAAAAAAAAAAAAAAAA==$/yMKBHC9e4Oj4Di5AOdOD+pciO7LV/uZPL7LWAvz7GY=",
		},
		{
			password: "qwerty🌸",
			salt: []byte{
				0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
				0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
			},
			want: "argon2id$v=19,t=4,m=65536,p=4$AAECAwQFBgcICQoLDA0ODw==$ejJPoTAtyy02N65TG7pPKfaZwLml7ZdePVPRG6twu2k=",
		},
		{
			password: "p4Mxd4t$j#*hAW6ZTw!uFD^ihS5s9i",
			salt: []byte{
				0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
				0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
			},
			want: "argon2id$v=19,t=4,m=65536,p=4$AAECAwQFBgcICQoLDA0ODw==$Ue8NB/5FMU22moAri/Ly7EHevWiuzY+eWM/SrpdF4go=",
		},
		{
			password: "hUbrDqTLWW&XghuDNfJv^$zXigE&8bJ8cqbrCfZ%W5Uwn9A2NWdwgsg@@QCgTik6$PxcYoGSF&6i!j%56r5Rdc*spnUcZTuC$NLKe4RtutK^k7LK&Q4MoP9eXfj54Kjh",
			salt: []byte{
				0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
				0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
			},
			want: "argon2id$v=19,t=4,m=65536,p=4$AAECAwQFBgcICQoLDA0ODw==$1nngZB8slAPvCLxF1mf0n3aR+CQmTWFkx4DXQMIV+QM=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			assert.Equal(t, tt.want, hashPasswd(tt.password, tt.salt))
		})
	}
}

func TestVerifyPasswd(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		hash      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "argon2id",
			password:  "qwerty🌸",
			hash:      "argon2id$v=19,t=4,m=65536,p=4$AAECAwQFBgcICQoLDA0ODw==$ejJPoTAtyy02N65TG7pPKfaZwLml7ZdePVPRG6twu2k=",
			assertion: assert.NoError,
		},
		{
			name:      "argon2id different options",
			password:  "qwerty🌸",
			hash:      "argon2id$v=19,t=8,m=32768,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$JMhcCrhKMFOtY5rcEmnAiDKw71ooKOGwIaeermvmouw=",
			assertion: assert.NoError,
		},
		{
			name:      "argon2id bad password",
			password:  "qwerty🌸 ",
			hash:      "argon2id$v=19,t=8,m=32768,p=4$9oV2HEQvNcSpxjkNBSAQAQ==$JMhcCrhKMFOtY5rcEmnAiDKw71ooKOGwIaeermvmouw=",
			assertion: assert.Error,
		},
		{
			name:      "nologin user",
			password:  "qwerty🌸 ",
			hash:      "nologin",
			assertion: assert.Error,
		},
		{
			name:      "unsupported algorithm",
			password:  "qwerty🌸 ",
			hash:      "sha512$def$asd",
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, VerifyPasswd(tt.password, tt.hash))
		})
	}
}

func TestHashAndVerifyPasswd(t *testing.T) {
	tests := []string{
		"qwerty",
		"qwerty🌸",
		"p4Mxd4t$j#*hAW6ZTw!uFD^ihS5s9i",
		"hUbrDqTLWW&XghuDNfJv^$zXigE&8bJ8cqbrCfZ%W5Uwn9A2NWdwgsg@@QCgTik6$PxcYoGSF&6i!j%56r5Rdc*spnUcZTuC$NLKe4RtutK^k7LK&Q4MoP9eXfj54Kjh",
	}
	for _, passwd := range tests {
		t.Run(passwd, func(t *testing.T) {
			hash, err := HashPasswd(passwd)
			assert.NoError(t, err)
			err = VerifyPasswd(passwd, hash)
			assert.NoError(t, err)
		})
	}
}

func BenchmarkHashPasswd(b *testing.B) {
	const password = "9NdKaUt4y%STB#!5tWX5"
	var err error
	for i := 0; i < b.N; i++ {
		_, err = HashPasswd(password)
		require.NoError(b, err)
	}
}
