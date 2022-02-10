package util

import (
	"encoding/json"
	"errors"
	"io"

	naclbox "golang.org/x/crypto/nacl/box"
)

func OpenJsonBox(data interface{}, box []byte, nonce *[24]byte, peersPublicKey, privateKey *[32]byte) error {
	out, ok := naclbox.Open(nil, box, nonce, peersPublicKey, privateKey)
	if !ok {
		return errors.New("unable to open box")
	}
	return json.Unmarshal(out, data)
}

func SealJsonBox(data interface{}, nonce *[24]byte, peersPublicKey, privateKey *[32]byte) ([]byte, error) {
	in, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return naclbox.Seal(nil, in, nonce, peersPublicKey, privateKey), nil
}

func GenerateNonce(rand io.Reader) (*[24]byte, error) {
	nonce := new([24]byte)
	_, err := io.ReadFull(rand, nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce, nil
}
