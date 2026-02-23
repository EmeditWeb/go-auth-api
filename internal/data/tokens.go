package data

import (
	"crypto/sha256"
	"time"
	"crypto/rand"
	"encoding/base32"
)

// defines the scope of the token
const (
	ScopeAuthentication = "authentication"
)

// the token struct
type Token struct {
	Plaintext string    `json:"token"`
    Hash      []byte    `json:"-"`
    UserID    int64     `json:"-"`
    Expiry    time.Time `json:"expiry"`
    Scope     string    `json:"-"`
}

// generateToken creates a new token
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	
	token := Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope: scope,
	}

	// generate random bytes for the token
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// encode the random bytes to a base32 string
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// hash the plaintext token
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return &token, nil
}
