package core

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	addErr   error
	statsErr error
	dropErr  error
	idsErr   error
}

func (m *mockDB) Add(_ context.Context, _ Comics) error {
	return m.addErr
}
func (m *mockDB) Stats(_ context.Context) (DBStats, error) {
	return DBStats{}, m.statsErr
}
func (m *mockDB) Drop(_ context.Context) error {
	return m.dropErr
}
func (m *mockDB) IDs(_ context.Context) ([]int, error) {
	return nil, m.idsErr
}

type mockXKCD struct {
	getErr    error
	lastIDErr error
	lastID    int
}

func (m *mockXKCD) Get(_ context.Context, _ int) (XKCDInfo, error) {
	return XKCDInfo{}, m.getErr
}
func (m *mockXKCD) LastID(_ context.Context) (int, error) {
	return m.lastID, m.lastIDErr
}

type mockWords struct {
	err error
}

func (m *mockWords) Norm(_ context.Context, _ string) ([]string, error) {
	return nil, m.err
}

type mockPublisher struct {
	err error
}

func (m *mockPublisher) Publish(_ context.Context) error {
	return m.err
}
func (m *mockPublisher) PublishDrop(_ context.Context) error {
	return m.err
}

func TestNewService_InvalidConcurrency(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{},
		&mockWords{},
		-1,
		&mockPublisher{},
	)
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewService_Valid(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestService_Status(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	assert.Equal(t, StatusIdle, svc.Status(context.Background()))
}

func TestService_Stats(t *testing.T) {
	tests := []struct {
		name    string
		dbErr   error
		xkcdErr error
		wantErr bool
	}{
		{
			name:    "db_error",
			dbErr:   ErrInternalServerError,
			wantErr: true,
		},
		{
			name:    "xkcd_error",
			xkcdErr: ErrInternalServerError,
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
				&mockDB{statsErr: tt.dbErr},
				&mockXKCD{lastIDErr: tt.xkcdErr},
				&mockWords{},
				1,
				&mockPublisher{},
			)
			require.NoError(t, err)
			_, err = svc.Stats(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestService_Drop(t *testing.T) {
	tests := []struct {
		name       string
		dropErr    error
		publishErr error
		wantErr    bool
	}{
		{
			name:    "db_error",
			dropErr: ErrInternalServerError,
			wantErr: true,
		},
		{
			name:       "publish_error",
			publishErr: ErrInternalServerError,
			wantErr:    true,
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
				&mockDB{dropErr: tt.dropErr},
				&mockXKCD{},
				&mockWords{},
				1,
				&mockPublisher{err: tt.publishErr},
			)
			require.NoError(t, err)
			err = svc.Drop(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Update_AlreadyRunning(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	svc.running.Store(true)
	err = svc.Update(context.Background())
	assert.ErrorIs(t, err, ErrAlreadyExists)
}

func TestService_Update_LastIDError(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{lastIDErr: ErrInternalServerError},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	err = svc.Update(context.Background())
	assert.Error(t, err)
}

func TestService_Update_DBIDsError(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{idsErr: ErrInternalServerError},
		&mockXKCD{lastID: 0},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	err = svc.Update(context.Background())
	assert.Error(t, err)
}

func TestService_Update_OK(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{lastID: 0},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	err = svc.Update(context.Background())
	assert.NoError(t, err)
}

func TestService_Update_WithComics(t *testing.T) {
	svc, err := NewService(
		slog.Default(),
		&mockDB{},
		&mockXKCD{lastID: 3},
		&mockWords{},
		1,
		&mockPublisher{},
	)
	require.NoError(t, err)
	err = svc.Update(context.Background())
	assert.NoError(t, err)
}
