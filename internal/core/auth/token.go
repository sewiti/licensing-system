package auth

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	cryptorand "crypto/rand"

	"github.com/vk-rv/pvx"
)

var ErrInvalidToken = errors.New("invalid token")

const TokenValidFor = 12 * time.Hour

type TokenManager struct {
	proto  *pvx.ProtoV4Public
	pk     *pvx.AsymPublicKey
	sk     *pvx.AsymSecretKey
	issuer string
}

func NewTokenManager(seed []byte, issuer string) (*TokenManager, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid seed length, expected %d", ed25519.SeedSize)
	}
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seed))
	if err != nil {
		return nil, err
	}
	return &TokenManager{
		proto:  pvx.NewPV4Public(),
		pk:     pvx.NewAsymmetricPublicKey(pub, pvx.Version4),
		sk:     pvx.NewAsymmetricSecretKey(priv, pvx.Version4),
		issuer: issuer,
	}, nil
}

func (t *TokenManager) IssueToken(subject string) (string, error) {
	return t.issueToken(cryptorand.Reader, subject, time.Now(), TokenValidFor)
}

func (t *TokenManager) IssueTokenIndefinite(subject string) (string, error) {
	return t.issueToken(cryptorand.Reader, subject, time.Now(), -1)
}

func (t *TokenManager) issueToken(rand io.Reader, subject string, now time.Time, validFor time.Duration) (string, error) {
	tokenID, err := generateTokenID(rand)
	if err != nil {
		return "", err
	}

	claims := pvx.RegisteredClaims{
		Issuer:   t.issuer,
		Subject:  subject,
		IssuedAt: pvx.TimePtr(now.UTC()),
		TokenID:  tokenID,
	}
	if validFor >= 0 {
		claims.Expiration = pvx.TimePtr(now.Add(validFor).UTC())
	}
	return t.proto.Sign(t.sk, &claims)
}

// Returns ErrInvalidToken
func (t *TokenManager) VerifyToken(token string) (subject string, err error) {
	claims, err := t.verifyToken(token)
	if err != nil {
		switch {
		case errors.Is(err, pvx.ErrInvalidSignature):
			return "", ErrInvalidToken
		case errors.Is(err, pvx.ErrMalformedToken):
			return "", ErrInvalidToken
		default:
			return "", err
		}
	}
	return claims.Subject, nil
}

func (t *TokenManager) verifyToken(token string) (*pvx.RegisteredClaims, error) {
	tok := t.proto.Verify(token, t.pk)
	var claims pvx.RegisteredClaims
	return &claims, tok.ScanClaims(&claims)
}

func generateTokenID(rand io.Reader) (string, error) {
	const idLen = 8
	token := make([]byte, idLen)
	_, err := io.ReadFull(rand, token)
	if err != nil {
		return "", fmt.Errorf("token-id: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(token), nil
}
