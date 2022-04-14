package util

import (
	"encoding/json"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
	naclbox "golang.org/x/crypto/nacl/box"
)

func OpenJsonBox(data interface{}, box []byte, nonce, peersPublicKey, privateKey []byte) error {
	n, pub, priv, err := nonceAndKeys(nonce, peersPublicKey, privateKey)
	if err != nil {
		return err
	}
	out, ok := naclbox.Open(nil, box, n, pub, priv)
	if !ok {
		return errors.New("unable to open box")
	}
	return json.Unmarshal(out, data)
}

func SealJsonBox(data interface{}, nonce, peersPublicKey, privateKey []byte) ([]byte, error) {
	n, pub, priv, err := nonceAndKeys(nonce, peersPublicKey, privateKey)
	if err != nil {
		return nil, err
	}
	in, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return naclbox.Seal(nil, in, n, pub, priv), nil
}

func GenerateNonce(rand io.Reader) ([]byte, error) {
	nonce := make([]byte, 24)
	_, err := io.ReadFull(rand, nonce)
	return nonce, err
}

func GenerateKey(rand io.Reader) (publicKey, privateKey []byte, err error) {
	pub, priv, err := box.GenerateKey(rand)
	if err != nil {
		return nil, nil, err
	}
	return pub[:], priv[:], nil
}

func Nonce(bs []byte) (*[24]byte, error) {
	if len(bs) != 24 {
		return nil, errors.New("invalid nonce length")
	}
	return (*[24]byte)(bs), nil
}

func Key(bs []byte) (*[32]byte, error) {
	if len(bs) != 32 {
		return nil, errors.New("invalid key length")
	}
	return (*[32]byte)(bs), nil
}

func nonceAndKeys(nonce, publicKey, privateKey []byte) (*[24]byte, *[32]byte, *[32]byte, error) {
	n, err := Nonce(nonce)
	if err != nil {
		return nil, nil, nil, err
	}
	pub, err := Key(publicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	priv, err := Key(privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	return n, pub, priv, nil
}
