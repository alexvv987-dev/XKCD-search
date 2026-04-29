package middleware

import (
	"net/http"
	"strings"
)

type TokenVerifier interface {
	Verify(token string) error
}

func Auth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Token ")
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err := verifier.Verify(token); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
