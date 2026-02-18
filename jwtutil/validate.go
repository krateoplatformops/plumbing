package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token is expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

func Validate(signingKey, bearer string) (UserInfo, error) {
	if signingKey == "" {
		return UserInfo{}, fmt.Errorf("signing key cannot be empty")
	}

	tok, err := jwt.ParseWithClaims(bearer, &KrateoClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(signingKey), nil
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
