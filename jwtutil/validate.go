package jwtutil

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token is expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// Validate parses and verifies the given bearer token against the supplied
// RSA public key. Only the RS256 asymmetric algorithm is accepted; tokens
// signed with any other method (in particular symmetric HMAC) are rejected to
// prevent algorithm-confusion attacks.
func Validate(publicKey *rsa.PublicKey, bearer string) (UserInfo, error) {
	if publicKey == nil {
		return UserInfo{}, fmt.Errorf("public key cannot be nil")
	}

	tok, err := jwt.ParseWithClaims(bearer, &KrateoClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			return UserInfo{}, ErrTokenInvalid
		}
		return UserInfo{}, ErrTokenExpired
	}

	claims, ok := tok.Claims.(*KrateoClaims)
	if !ok {
		return UserInfo{}, ErrTokenInvalid
	}

	return claims.UserInfo, nil
}
