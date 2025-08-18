package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type PKCEChallenge struct {
	Verifier  string
	Challenge string
}

// GeneratePKCEChallenge generates a PKCE code verifier and challenge
func GeneratePKCEChallenge() (*PKCEChallenge, error) {
	// Generate code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	verifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(verifierBytes)

	// Generate code challenge (SHA256 hash of verifier)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
	}, nil
}
