package util

import (
	"encoding/base64"
	"testing"

	cryptorand "crypto/rand"

	"github.com/stretchr/testify/require"
)

func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateKey(cryptorand.Reader)
		require.NoError(b, err)
	}
}

func BenchmarkGenerateNonce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateNonce(cryptorand.Reader)
		require.NoError(b, err)
	}
}

func BenchmarkSealJsonBox(b *testing.B) {
	data := struct {
		Benchmark string `json:"benchmark"`
	}{
		Benchmark: "go benchmark testing",
	}
	nonce := mustParseBase64("C0zaX2YG0mRZMD9AYCgKOs4AQjMT/JaR")
	public := mustParseBase64("WxTr9I1THoayJrSalVa6Y5xYThUydWIx0RMwcsadxiA=")
	private := mustParseBase64("KaxvhSWBy24qXM0NwPxfICuI8q4lidLsB+xnRZ/H9m8=")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SealJsonBox(data, nonce, public, private)
		require.NoError(b, err)
	}
}

func BenchmarkOpenJsonBox(b *testing.B) {
	var data struct {
		Benchmark string `json:"benchmark"`
	}
	box := mustParseBase64("7iHash7fha2GvoPv4jZsjd3xZp53iP4xQk1fHZxONYEXcOA6SBXik/j6GtZEDdA2tsDkkw==")
	nonce := mustParseBase64("C0zaX2YG0mRZMD9AYCgKOs4AQjMT/JaR")
	public := mustParseBase64("JurTC5R3kvKGqQWuZSa8b9Zp/H1B+23oHPgGRfKMxR8=")
	private := mustParseBase64("pfVqr5P1om+NQ00oWV5sNqyZtjnxbyXAkEqvVOswFnM=")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := OpenJsonBox(&data, box, nonce, public, private)
		require.NoError(b, err)
	}
}

func mustParseBase64(s string) []byte {
	bs, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bs
}
