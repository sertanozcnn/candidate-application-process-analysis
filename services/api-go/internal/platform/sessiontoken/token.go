package sessiontoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const tokenLength = 32

func Generate() (string, error) {
	buf := make([]byte, tokenLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
