package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	cryptorand "crypto/rand"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidPasswd = errors.New("invalid password")
	ErrNoLogin       = errors.New("no login")
)

const hashSeparator = "$"

func HashPasswd(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := cryptorand.Read(salt)
	if err != nil {
		return "", err
	}
	return hashPasswd(password, salt), nil
}

func hashPasswd(password string, salt []byte) string {
	const (
		optAlg      = "argon2id"
		optTime     = 4
		optMem      = 64 * 1024 // 64 MiB in KiB
		optParallel = 4
	)
	hash := argon2.IDKey([]byte(password), salt, optTime, optMem, optParallel, 32)
	params := fmt.Sprintf("v=%d,t=%d,m=%d,p=%d", argon2.Version, optTime, optMem, optParallel)
	data := []string{
		optAlg,
		params,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(hash),
	}
	return strings.Join(data, hashSeparator)
}

// VerifyPasswd
//
// Returns ErrInvalidPassword
// Returns ErrNoLogin
func VerifyPasswd(password, hash string) error {
	parts := strings.Split(hash, hashSeparator)
	if len(parts) == 0 {
		return fmt.Errorf("auth: missing algorithm: %s", hash)
	}

	switch alg := parts[0]; alg {
	case "nologin":
		return fmt.Errorf("auth: %w", ErrNoLogin)

	case "argon2id":
		err := verifyArgon2id(password, parts)
		if err != nil {
			return fmt.Errorf("auth: argon2id: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("auth: unsupported algorithm %s: %s", alg, hash)
	}
}

func verifyArgon2id(password string, hashParts []string) error {
	if len(hashParts) != 4 {
		return fmt.Errorf("expected 4 hash parts: %s", hashParts)
	}
	var version int
	var optTime uint32
	var optMem uint32
	var optParallel uint8
	_, err := fmt.Sscanf(hashParts[1], "v=%d,t=%d,m=%d,p=%d", &version, &optTime, &optMem, &optParallel)
	if err != nil {
		return fmt.Errorf("options: %w", err)
	}

	if version != argon2.Version {
		return errors.New("incompatible version")
	}
	salt, err := base64.StdEncoding.DecodeString(hashParts[2])
	if err != nil {
		return fmt.Errorf("salt: %w", err)
	}
	passwdHash := argon2.IDKey([]byte(password), salt, optTime, optMem, optParallel, 32)
	if base64.StdEncoding.EncodeToString(passwdHash) != hashParts[3] {
		return ErrInvalidPasswd
	}
	return nil
}
