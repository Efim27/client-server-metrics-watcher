package middleware

import (
	"bytes"
	"crypto/rsa"
	"io/ioutil"
	"net/http"

	handlerRSA "metrics/internal/rsa"
)

func NewRSAHandle(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey != nil {
				next.ServeHTTP(w, r)
				return
			}

			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			decryptedBody := handlerRSA.DecryptWithPrivateKey(bodyBytes, privateKey)
			r.Body = ioutil.NopCloser(bytes.NewReader(decryptedBody))

			next.ServeHTTP(w, r)
		})
	}
}
