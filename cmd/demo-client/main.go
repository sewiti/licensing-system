package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/sewiti/licensing-system/pkg/license"
)

func main() {
	var licenseKeyStr string
	var serverIDStr string
	var machineIDFile string
	var url string
	var maxRefresh time.Duration
	var n int
	flag.StringVar(&licenseKeyStr, "license-key", "", "License key.")
	flag.StringVar(&serverIDStr, "server-id", "", "Licensing server ID key (public).")
	flag.StringVar(&machineIDFile, "machine-id-file", "/etc/machine-id", "Machine ID file.")
	flag.StringVar(&url, "url", "http://localhost/api/license-sessions", "Licensing server sessions endpoint url.")
	flag.DurationVar(&maxRefresh, "max-refresh", time.Minute, "Maximum refresh time, useful for responsive demo.")
	flag.IntVar(&n, "instances", 1, "Number of license sessions.")
	flag.Parse()

	if licenseKeyStr == "" {
		exitf(1, "missing license-key\n")
	}
	licenseKeyStr = strings.TrimSuffix(licenseKeyStr, "=")
	licenseKey, err := base64.RawStdEncoding.DecodeString(licenseKeyStr)
	if err != nil {
		exitf(1, "license-key: %v\n", err)
	}

	if serverIDStr == "" {
		exitf(1, "missing server-id\n")
	}
	serverIDStr = strings.TrimSuffix(serverIDStr, "=")
	serverID, err := base64.RawStdEncoding.DecodeString(serverIDStr)
	if err != nil {
		exitf(1, "server-id: %v\n", err)
	}

	machineID, err := license.ReadID(machineIDFile)
	if err != nil {
		exitf(1, "machine-id: %v\n", err)
	}

	wg := sync.WaitGroup{}
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg.Add(2 * n)
	for i := 0; i < n; i++ {
		cl, err := license.NewClient(url, serverID, machineID, licenseKey)
		if err != nil {
			exitf(1, "%v\n", err)
		}

		go func(i int) {
			defer wg.Done()
			cl.Run(ctx, maxRefresh, func(msg string, err error) {
				if err != nil {
					log.WithError(err).Errorf("%d: %s", i, msg)
					return
				}
				log.Infof("%d: %s", i, msg)
			})
		}(i)

		go func(i int) {
			defer wg.Done()
			t := time.NewTicker(maxRefresh)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					productName, err := cl.ProductName()
					if err != nil {
						log.WithError(err).Errorf("%d: state: %v", i, cl.State())
						continue
					}
					productData, err := cl.ProductData()
					if err != nil {
						log.WithError(err).Errorf("%d: state: %v", i, cl.State())
						continue
					}
					data, err := cl.Data()
					if err != nil {
						log.WithError(err).Errorf("%d: state: %v", i, cl.State())
						continue
					}
					log.Infof("%d: state: %v; product-name: %s; product-data: %s; license-data: %s", i, cl.State(), productName, productData, data)
				}
			}
		}(i)
	}
	<-ctx.Done()
	wg.Wait()
}

func exitf(code int, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(code)
}
