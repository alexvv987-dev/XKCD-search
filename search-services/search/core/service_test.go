package core

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	err    error
	result []Comics
}

func (m *mockDB) Get(_ context.Context) ([]Comics, error) {
	return m.result, m.err
}
func (m *mockDB) GetByIDs(_ context.Context, _ []int) ([]Comics, error) {
	return m.result, m.err
}
func (m *mockDB) Search(_ context.Context, _ []string) ([]Comics, error) {
	return m.result, m.err
}

type mockWords struct {
	err    error
	result []string
}

func (m *mockWords) Norm(_ context.Context, _ string) ([]string, error) {
	return m.result, m.err
}

func TestNewService_InvalidConcurrency(t *testing.T) {
	svc, err := NewService(slog.Default(), &mockDB{}, &mockWords{}, -1)
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewService_Valid(t *testing.T) {
	svc, err := NewService(slog.Default(), &mockDB{}, &mockWords{}, 1)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_Index_DBError(t *testing.T) {
	svc, err := NewService(slog.Default(), &mockDB{err: ErrNotFound}, &mockWords{}, 1)
	require.NoError(t, err)
	err = svc.Index(context.Background())
	assert.Error(t, err)
}

func TestService_Index_OK(t *testing.T) {
	svc, err := NewService(slog.Default(), &mockDB{}, &mockWords{}, 1)
	require.NoError(t, err)
	err = svc.Index(context.Background())
	assert.NoError(t, err)
}

func TestService_ResetIndex(t *testing.T) {
	svc, err := NewService(slog.Default(), &mockDB{}, &mockWords{}, 1)
	require.NoError(t, err)
	err = svc.ResetIndex(context.Background())
	assert.NoError(t, err)
}

func TestService_Search(t *testing.T) {
	tests := []struct {
		name     string
		dbErr    error
		wordsErr error
		wantErr  bool
	}{
		{
			name:     "words_error",
			wordsErr: ErrInternalServerError,
			wantErr:  true,
		},
		{
			name:    "db_error",
			dbErr:   ErrInternalServerError,
			wantErr: true,
		},
		{
			name:    "ok",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewService(
				slog.Default(),
				&mockDB{err: tt.dbErr},
				&mockWords{err: tt.wordsErr},
				1,
			)
			require.NoError(t, err)
			_, err = svc.Search(context.Background(), "linux", 10)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SearchIndex(t *testing.T) {
	tests := []struct {
		name     string
		dbErr    error
		wordsErr error
		wantErr  bool
	}{
		{
			name:     "words_error",
			wordsErr: ErrInternalServerError,
			wantErr:  true,
		},
		{
			name:    "db_error",
			dbErr:   ErrInternalServerError,
			wantErr: true,
		},
		{
			name:    "ok",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewService(
				slog.Default(),
				&mockDB{err: tt.dbErr},
				&mockWords{err: tt.wordsErr},
				1,
			)
			require.NoError(t, err)
			_, err = svc.SearchIndex(context.Background(), "linux", 10)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
