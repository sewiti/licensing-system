package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"

	cryptorand "crypto/rand"

	"golang.org/x/crypto/nacl/box"
)

func generateKeys(args []string) error {
	var fHex bool
	var fBase64 bool
	fs := flag.NewFlagSet("generate-keys", flag.ExitOnError)
	fs.BoolVar(&fHex, "hex", false, "Print keys hex-encoded.")
	fs.BoolVar(&fBase64, "base64", false, "Print keys base64 encoded.")
	fs.Parse(args)

	pub, priv, err := box.GenerateKey(cryptorand.Reader)
	if err != nil {
		return err
	}

	if fHex || !fBase64 {
		fmt.Printf("id:hex:%s\n", hex.EncodeToString(pub[:]))
	}
	if !fHex || fBase64 {
		fmt.Printf("id:base64:%s\n", base64.StdEncoding.EncodeToString(pub[:]))
	}

	if fHex || !fBase64 {
		fmt.Printf("key:hex:%s\n", hex.EncodeToString(priv[:]))
	}
	if !fHex || fBase64 {
		fmt.Printf("key:base64:%s\n", base64.StdEncoding.EncodeToString(priv[:]))
	}
	return nil
}
