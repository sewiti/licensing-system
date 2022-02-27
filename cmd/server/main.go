package main

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/nacl/box"
)

func main() {
	if len(os.Args) <= 1 {
		runServer()
		return
	}
	switch os.Args[1] {
	case "generate-keys":
		err := generateKeys()
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate-keys: %v\n", err)
			os.Exit(1)
		}

	case "-h", "--help":
		printUsage(os.Stdout)

	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "usage: %s [generate-keys]\n", os.Args[0])
}

func generateKeys() error {
	pub, priv, err := box.GenerateKey(cryptorand.Reader)
	if err != nil {
		return err
	}
	fmt.Println("id (public):")
	fmt.Printf("  hex:    % #x\n", pub[:])
	fmt.Printf("  base64: %s\n", base64.StdEncoding.EncodeToString(pub[:]))
	fmt.Println()
	fmt.Println("key (private):")
	fmt.Printf("  hex:    % #x\n", priv[:])
	fmt.Printf("  base64: %s\n", base64.StdEncoding.EncodeToString(priv[:]))
	return nil
}
