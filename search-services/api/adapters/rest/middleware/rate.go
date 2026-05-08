package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

const (
	rateLimitBurst = 1
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	if rps <= 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "rate limit misconfigured", http.StatusServiceUnavailable)
		}
	}
	limiter := rate.NewLimiter(rate.Limit(rps), rateLimitBurst)
	return func(w http.ResponseWriter, r *http.Request) {
		if err := limiter.Wait(r.Context()); err != nil {
			return
		}
		next(w, r)
	}
}
