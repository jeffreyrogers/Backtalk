package sessions

import (
	"crypto/rand"
	"crypto/subtle"
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

// TODO: need to cryptographically sign the session key
func GenerateSessionKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Printf("error generating session key: %v\n", err)
		os.Exit(1)
	}

	return key
}

// TODO: implement
func SessionKeyValid(key []byte) bool {
	return true
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
