package security

import (
	"crypto/hmac"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func GenerateSalt() ([]byte, error) {
	return GetRandomBytes(16)
}

// Generate cryptographically random 32 byte string, sign with hmac, and return base64 encoded version
// This base64 string can be stored in a cookie on the client to keep track of sessions. It should also be
// stored in the DB.
func GenerateSessionKey() ([]byte, string) {
	sessionID, err := GetRandomBytes(32)
	if err != nil {
		fmt.Printf("error generating session ID: %v\n", err)
		return nil, ""
	}

	signature := sign(sessionID)
	rawSessionKey := append(sessionID, signature...)
	return sessionID, base64.StdEncoding.EncodeToString(rawSessionKey)
}

func SessionIDValid(key string) ([]byte, bool) {
	decodedSessionKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Printf("Error decoding session key: %v\n", err)
		return nil, false
	}

	sessionID := decodedSessionKey[:32]
	signature := decodedSessionKey[32:]
	computedSignature := sign(sessionID)
	return sessionID, hmac.Equal(signature, computedSignature)
}

func Hash(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}

func PasswordValid(validHash []byte, password string, salt []byte) bool {
	hash := Hash(password, salt)
	return subtle.ConstantTimeCompare(validHash, hash) == 1
}
