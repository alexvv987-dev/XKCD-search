package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

func TestApi_toHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "unknown_error",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unauthorized",
			err:        core.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not_found",
			err:        core.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "internal_server_error",
			err:        core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "bad_request",
			err:        core.ErrBadArguments,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantStatus, toHTTPStatus(tt.err))
		})
	}
}

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestApi_PingHandler(t *testing.T) {
	tests := []struct {
		name      string
		pingerErr error
		wantReply string
	}{
		{
			name:      "pinger_error",
			pingerErr: errors.New("unavailable"),
			wantReply: "unavailable",
		},
		{
			name:      "pinger_ok",
			pingerErr: nil,
			wantReply: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pingers := map[string]core.Pinger{
				"test": &mockPinger{err: tt.pingerErr},
			}
			handler := NewPingHandler(slog.Default(), pingers)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			handler(w, r)
			var resp PingResponse
			require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
			assert.Equal(t, tt.wantReply, resp.Replies["test"])
		})

	}
}

type mockSearcher struct {
	err    error
	result []core.Comics
}

func (m *mockSearcher) Search(_ context.Context, _ string, _ int) ([]core.Comics, error) {
	return m.result, m.err
}
func (m *mockSearcher) SearchIndex(_ context.Context, _ string, _ int) ([]core.Comics, error) {
	return m.result, m.err
}

func TestApi_SearchHandler(t *testing.T) {
	tests := []struct {
		name       string
		phrase     string
		limit      string
		searchErr  error
		wantStatus int
	}{
		{
			name:       "empty_phrase",
			phrase:     "",
			limit:      "",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_limit_alpha",
			phrase:     "linux",
			limit:      "abc",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_limit_negative",
			phrase:     "linux",
			limit:      "-1",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "searcher_error",
			phrase:     "linux",
			limit:      "10",
			searchErr:  core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success",
			phrase:     "linux",
			limit:      "10",
			searchErr:  nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := &mockSearcher{err: tt.searchErr}
			handler := NewSearchHandler(slog.Default(), searcher)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			q := r.URL.Query()
			q.Set("phrase", tt.phrase)
			q.Set("limit", tt.limit)
			r.URL.RawQuery = q.Encode()
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
func TestApi_SearchIndexHandler(t *testing.T) {
	tests := []struct {
		name       string
		phrase     string
		limit      string
		searchErr  error
		wantStatus int
	}{
		{
			name:       "empty_phrase",
			phrase:     "",
			limit:      "",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_limit_alpha",
			phrase:     "linux",
			limit:      "abc",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_limit_negative",
			phrase:     "linux",
			limit:      "-1",
			searchErr:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "searcher_error",
			phrase:     "linux",
			limit:      "10",
			searchErr:  core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success",
			phrase:     "linux",
			limit:      "10",
			searchErr:  nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := &mockSearcher{err: tt.searchErr}
			handler := NewSearchIndexHandler(slog.Default(), searcher)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			q := r.URL.Query()
			q.Set("phrase", tt.phrase)
			q.Set("limit", tt.limit)
			r.URL.RawQuery = q.Encode()
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

type mockUpdater struct {
	err    error
	stats  core.UpdateStats
	status core.UpdateStatus
}

func (m *mockUpdater) Update(_ context.Context) error {
	return m.err
}

func (m *mockUpdater) Stats(_ context.Context) (core.UpdateStats, error) {
	return m.stats, m.err
}

func (m *mockUpdater) Status(_ context.Context) (core.UpdateStatus, error) {
	return m.status, m.err
}

func (m *mockUpdater) Drop(_ context.Context) error {
	return m.err
}

func TestApi_UpdateHandler(t *testing.T) {
	tests := []struct {
		name       string
		updaterErr error
		wantStatus int
	}{
		{
			name:       "error",
			updaterErr: core.ErrAlreadyExists,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "ok",
			updaterErr: nil,
			wantStatus: http.StatusAccepted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{err: tt.updaterErr}
			handler := NewUpdateHandler(slog.Default(), updater)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestApi_UpdateStatsHandler(t *testing.T) {
	tests := []struct {
		name       string
		updaterErr error
		wantStatus int
	}{
		{name: "error", updaterErr: core.ErrInternalServerError, wantStatus: http.StatusInternalServerError},
		{name: "ok", updaterErr: nil, wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{err: tt.updaterErr}
			handler := NewUpdateStatsHandler(slog.Default(), updater)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestApi_UpdateStatusHandler(t *testing.T) {
	tests := []struct {
		name       string
		updaterErr error
		wantStatus int
	}{
		{
			name:       "error",
			updaterErr: core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "ok",
			updaterErr: nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{err: tt.updaterErr}
			handler := NewUpdateStatusHandler(slog.Default(), updater)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestApi_DropHandler(t *testing.T) {
	tests := []struct {
		name       string
		updaterErr error
		wantStatus int
	}{
		{
			name:       "error",
			updaterErr: core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "ok",
			updaterErr: nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{err: tt.updaterErr}
			handler := NewDropHandler(slog.Default(), updater)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/", nil)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

type mockNormalizer struct {
	err    error
	result []string
}

func (m *mockNormalizer) Norm(_ context.Context, _ string) ([]string, error) {
	return m.result, m.err
}

func TestApi_WordsHandler(t *testing.T) {
	tests := []struct {
		name       string
		normErr    error
		wantStatus int
	}{
		{
			name:       "error",
			normErr:    core.ErrInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "ok",
			normErr:    nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := &mockNormalizer{err: tt.normErr}
			handler := NewWordsHandler(slog.Default(), normalizer)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/?phrase=linux", nil)
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

type mockAuthenticator struct {
	err   error
	token string
}

func (m *mockAuthenticator) Login(_, _ string) (string, error) {
	return m.token, m.err
}

func TestApi_LoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		loginErr   error
		wantStatus int
	}{
		{
			name:       "invalid_json",
			body:       "invalid",
			loginErr:   nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "login_error",
			body:       `{"name":"u","password":"p"}`,
			loginErr:   core.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ok",
			body:       `{"name":"u","password":"p"}`,
			loginErr:   nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &mockAuthenticator{err: tt.loginErr, token: "token123"}
			handler := NewLoginHandler(slog.Default(), auth)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			handler(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
