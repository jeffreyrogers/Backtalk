package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"

	"github.com/jeffreyrogers/backtalk/internal/globals"
)

func GetRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func sign(message []byte) []byte {
	h := hmac.New(sha512.New512_256, globals.AuthKey)
	h.Write(message)
	return h.Sum(nil)
}

func encode(token []byte) string {
	signature := sign(token)
	signedToken := append(token, signature...)
	return base64.StdEncoding.EncodeToString(signedToken)
}

func decode(encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	token := decoded[:32]
	signature := decoded[32:]
	computedSignature := sign(token)
	if hmac.Equal(signature, computedSignature) {
		return token, nil
	}

	return nil, errors.New("Signatures do not match")
}
