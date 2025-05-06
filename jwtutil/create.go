package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JwtSecretEnvKey = "AUTHN_JWT_SECRET"

	defaultISS = "krateo.io"
)

type UserInfo struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

type KrateoClaims struct {
	UserInfo
	jwt.RegisteredClaims
}

type CreateTokenOptions struct {
	Username   string
	Groups     []string
	Duration   time.Duration
	SigningKey string
}

// CreateToken generates a signed JWT token using the provided
// username, group list, and expiration duration.
// The token is signed using the HS256 algorithm.
// The signing key is read from the environment variable AUTHN_JWT_SECRET.
// If the environment variable is not set, the function returns an error.
func CreateToken(opts CreateTokenOptions) (string, error) {
	if opts.SigningKey == "" {
		return "", fmt.Errorf("signing key cannot be empty")
	}

	if opts.Username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	now := time.Now()

	claims := KrateoClaims{
		UserInfo: UserInfo{
			Username: opts.Username,
			Groups:   opts.Groups,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    defaultISS,
			ExpiresAt: jwt.NewNumericDate(now.Add(opts.Duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   opts.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(opts.SigningKey))
}
