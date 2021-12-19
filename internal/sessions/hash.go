package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"
)

func GenerateSalt() []byte {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		fmt.Printf("error generating salt: %v\n", err)
		os.Exit(1)
	}

	return salt
}

func sign(message []byte) []byte {
	hmacKey := os.Getenv("BACKTALK_AUTH_KEY")
	h := hmac.New(sha512.New512_256, []byte(hmacKey))
	h.Write(message)
	return h.Sum(nil)
}

// Generate cryptographically random 32 byte string, sign with hmac, and return base64 encoded version
// This base64 string can be stored in a cookie on the client to keep track of sessions. It should also be
// stored in the DB.
func GenerateSessionKey() string {
	sessionID := make([]byte, 32)
	_, err := rand.Read(sessionID)
	if err != nil {
		fmt.Printf("error generating session ID: %v\n", err)
		os.Exit(1)
	}

	signature := sign(sessionID)
	// TODO: check if there is a better way to concatenate two byte slices
	rawSessionKey := append(sessionID, signature...)
	return base64.StdEncoding.EncodeToString(rawSessionKey)
}

func SessionIDValid(key string) ([]byte, bool) {
	decodedSessionKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Printf("Error decoding session key: %v\n", err)
		os.Exit(1)
	}

	sessionID := decodedSessionKey[:32]
	signature := decodedSessionKey[32:]
	computedSignature := sign(sessionID)
	return sessionID, hmac.Equal(signature, computedSignature)
}

func Hash(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

func PasswordValid(validHash, password, salt []byte) bool {
	hash := Hash(password, salt)
	if subtle.ConstantTimeCompare(validHash, hash) == 1 {
		return true
	} else {
		return false
	}
}
