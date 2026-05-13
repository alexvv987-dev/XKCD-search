package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRate(t *testing.T) {
	tests := []struct {
		name       string
		rps        int
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "invalid_rps",
			rps:        0,
			wantStatus: http.StatusServiceUnavailable,
			wantNext:   false,
		},
		{
			name:       "valid_rps",
			rps:        100,
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			handler := Rate(next, tt.rps)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantNext, nextCalled)
		})
	}
}
