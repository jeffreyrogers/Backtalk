package csrf

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/jeffreyrogers/backtalk/internal/globals"
)

type csrf struct {
	h            http.Handler
	authKey      []byte
	errorHandler http.HandlerFunc
}

var CSRFTag = "csrf"
var fieldName = "csrf"
var cookieName = "csrf"
var tokenKey = "csrfToken"

var safeMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

func CSRFField(r *http.Request) template.HTML {
	input := fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, token(r))
	return template.HTML(input)
}

func Protect(authKey []byte) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &csrf{
			h:            h,
			authKey:      authKey,
			errorHandler: http.HandlerFunc(csrfFailure),
		}
	}
}

func (cs *csrf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(r)
	if err != nil {
		token, err := otp()
		if err != nil {
			cs.errorHandler.ServeHTTP(w, r)
			return
		}

		encoded := encode(token)

		c := &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Secure:   true,
			HttpOnly: true,
			MaxAge:   60 * 60 * 12,
			Expires:  time.Now().Add(time.Duration(60*60*12) * time.Second),
		}

		http.SetCookie(w, c)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, tokenKey, mask(token, r))
	r = r.WithContext(ctx)

	if !contains(safeMethods, r.Method) {
		// If X-Requested-By is set, then this cannot a CSRF attack because custom headers are not sent cross origin
		requestedBy := r.Header.Get("X-REQUESTED-BY")
		if requestedBy == "" {
			// Since X-Requested-By not set, we need to check the token sent in the request and see if it matches
			requestToken := r.PostFormValue(fieldName)
			if requestToken == "" {
				cs.errorHandler.ServeHTTP(w, r)
				return
			}

			maskedToken, err := base64.StdEncoding.DecodeString(requestToken)
			if err != nil {
				cs.errorHandler.ServeHTTP(w, r)
				return
			}

			unmaskedToken := unmask(maskedToken)
			if subtle.ConstantTimeCompare(unmaskedToken, token) != 1 {
				cs.errorHandler.ServeHTTP(w, r)
				return
			}
		}
	}

	w.Header().Add("Vary", "Cookie")
	cs.h.ServeHTTP(w, r)
}

func token(r *http.Request) string {
	if val := r.Context().Value(tokenKey); val != nil {
		if maskedToken, ok := val.(string); ok {
			return maskedToken
		}
	}

	return ""
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

func getToken(r *http.Request) ([]byte, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	return decode(cookie.Value)
}

func otp() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func mask(token []byte, r *http.Request) string {
	otp, err := otp()
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(append(otp, xor(otp, token)...))
}

func unmask(token []byte) []byte {
	if len(token) != 64 {
		return nil
	}

	otp := token[:32]
	masked := token[32:]

	return xor(otp, masked)
}

func xor(a, b []byte) []byte {
	n := min(len(a), len(b))
	res := make([]byte, n)

	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}

	return res
}

func contains(vals []string, s string) bool {
	for _, v := range vals {
		if v == s {
			return true
		}
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func csrfFailure(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("%s - CSRF Token or Header Invalid", http.StatusText(http.StatusForbidden)), http.StatusForbidden)
}
