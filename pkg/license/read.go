package license

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"os"
)

func ReadID(path string) ([]byte, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	bs = bytes.TrimSpace(bs)
	bs = bytes.ReplaceAll(bs, []byte{'-'}, nil)

	id := make([]byte, hex.DecodedLen(len(bs)))
	_, err = hex.Decode(id, bs)
	return id, err
}

func ReadKey(path string) ([]byte, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	bs = bytes.TrimSpace(bs)
	bs = bytes.TrimSuffix(bs, []byte{'='})

	key := make([]byte, base64.RawStdEncoding.DecodedLen(len(bs)))
	_, err = base64.RawStdEncoding.Decode(key, bs)
	return key, err
}
