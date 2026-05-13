package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockVerifier struct {
	err error
}

func (m *mockVerifier) Verify(_ string) error {
	return m.err
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		verifyErr  error
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "no_header",
			header:     "",
			verifyErr:  nil,
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "invalid_token",
			header:     "Token abc123",
			verifyErr:  errors.New("invalid token"),
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "valid_token",
			header:     "Token abc123",
			verifyErr:  nil,
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(tt.wantStatus)
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			handler := Auth(next, &mockVerifier{err: tt.verifyErr})
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantNext, nextCalled)
		})
	}
}
