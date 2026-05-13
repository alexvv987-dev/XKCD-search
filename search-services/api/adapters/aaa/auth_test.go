package aaa

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAAA_Login(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		password string
		wantErr  bool
	}{
		{
			name:     "user_not_found",
			user:     "",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "wrong_password",
			user:     "admin",
			password: "",
			wantErr:  true,
		},
		{
			name:     "valid_credentials",
			user:     "admin",
			password: "password",
			wantErr:  false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ADMIN_USER", "admin")
			t.Setenv("ADMIN_PASSWORD", "password")
			a, err := New(time.Minute, slog.Default())
			require.NoError(t, err)
			token, err := a.Login(tc.user, tc.password)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAAA_Verify_InvalidToken(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "password")
	a, err := New(time.Minute, slog.Default())
	require.NoError(t, err)
	err = a.Verify("invalid.token.string")
	assert.Error(t, err)
}

func TestAAA_Verify_ValidToken(t *testing.T) {
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "password")
	a, err := New(time.Minute, slog.Default())
	require.NoError(t, err)
	token, err := a.Login("admin", "password")
	require.NoError(t, err)
	err = a.Verify(token)
	assert.NoError(t, err)
}
