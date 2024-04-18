package secrets

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
)

type SecretHandler struct {
	Path   string
	secret []byte
}

// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
func (s SecretHandler) Validate(hash []byte, contents []byte) bool {
	hmac := hmac.New(sha256.New, s.secret)
	hmac.Write(contents)
	shasum := hmac.Sum(nil)
	return subtle.ConstantTimeCompare(hash, shasum) == 1
}

func New(path string, secret []byte) *SecretHandler {
	return &SecretHandler{path, secret}
}
