package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

func manageLicenseIssuer(args []string) error {
	var socket string
	fs := flag.NewFlagSet("issuer", flag.ExitOnError)
	fs.StringVar(&socket, "socket", "/run/licensing-server.sock", "Internal licensing server socket.")
	fs.Parse(args)

	args = fs.Args()
	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}
	username := args[0]
	action := args[1]
	if username == "" || strings.ContainsRune(username, '/') {
		return fmt.Errorf("invalid username: %s", username)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch action {
	case "enable", "disable":
		err := updateLicenseIssuerActive(ctx, socket, username, action == "enable")
		if err != nil {
			return err
		}
		fmt.Printf("%s has been %sd\n", username, action)
		return nil

	case "chpasswd":
		fmt.Print("New password: ")
		passwd, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return err
		}
		fmt.Println()
		err = updateLicenseIssuerPassword(ctx, socket, username, string(passwd))
		if err != nil {
			return err
		}
		fmt.Printf("%s password has been changed\n", username)
		return nil

	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

func updateLicenseIssuerActive(ctx context.Context, socket, username string, active bool) error {
	url := fmt.Sprintf("http://unix/license-issuers/%s/active", username)
	data := struct {
		Active bool `json:"active"`
	}{
		Active: active,
	}
	r, err := doInternalReq(ctx, socket, http.MethodPatch, url, data)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case 200, 204:
		return nil
	case 404:
		return fmt.Errorf("license issuer not found")
	default:
		return fmt.Errorf("unexpected status code: %s", r.Status)
	}
}

func updateLicenseIssuerPassword(ctx context.Context, socket, username, password string) error {
	url := fmt.Sprintf("http://unix/license-issuers/%s/change-password", username)
	data := struct {
		NewPassword string `json:"newPassword"`
	}{
		NewPassword: password,
	}
	r, err := doInternalReq(ctx, socket, http.MethodPatch, url, data)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case 200, 204:
		return nil
	case 400:
		msg, err := parseMessage(r.Body)
		if err != nil {
			return fmt.Errorf("invalid input")
		}
		return fmt.Errorf("invalid input: %s", msg)
	case 404:
		return fmt.Errorf("license issuer not found")
	default:
		return fmt.Errorf("unexpected status code: %s", r.Status)
	}
}

func doInternalReq(ctx context.Context, socket string, method, url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	cl := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := &net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				return d.DialContext(ctx, "unix", socket)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return cl.Do(req)
}

func parseMessage(body io.Reader) (string, error) {
	var msg struct {
		Message string `json:"message"`
	}
	err := json.NewDecoder(body).Decode(&msg)
	return msg.Message, err
}
