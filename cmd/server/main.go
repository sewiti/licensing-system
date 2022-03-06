package main

import (
	"fmt"
	"io"
	"os"

	"github.com/apex/log"
)

func main() {
	if len(os.Args) <= 1 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		err := runServer() // manages errors on it's own
		if err != nil {
			log.WithError(err).Fatal("run server")
		}

	case "generate-keys":
		err := generateKeys(os.Args[2:])
		if err != nil {
			log.WithError(err).Fatal("generate keys")
		}

	case "issuer":
		err := manageLicenseIssuer(os.Args[2:])
		if err != nil {
			log.WithError(err).Fatal("manage license issuer")
		}

	case "-h", "-help", "--help":
		printUsage(os.Stdout)

	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  %s <command> [arguments]\n", os.Args[0])
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  run                                                  run licensing server")
	fmt.Fprintln(w, "  issuer <username> enable|disable|chpasswd [-socket]  manage license issuers")
	fmt.Fprintln(w, "  generate-keys [-base64|-hex]                         generate random keys")
}
