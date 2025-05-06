package jwtutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractUserInfo extracts the "sub" and "groups" fields from a JWT payload
// without verifying the signature.
// Use only for non-security-critical use (e.g., logging).
func ExtractUserInfo(token string) (UserInfo, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return UserInfo{}, fmt.Errorf("invalid JWT format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return UserInfo{}, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return UserInfo{}, fmt.Errorf("invalid payload JSON: %w", err)
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return UserInfo{}, fmt.Errorf("subject (sub) claim not found")
	}

	groups := []string{}
	if rawGroups, ok := claims["groups"].([]any); ok {
		for _, g := range rawGroups {
			if str, ok := g.(string); ok {
				groups = append(groups, str)
			}
		}
	}

	return UserInfo{Username: sub, Groups: groups}, nil
}
