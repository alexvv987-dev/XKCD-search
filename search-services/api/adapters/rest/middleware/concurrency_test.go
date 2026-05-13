package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrency_SlotAvailable(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Concurrency(next, 1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConcurrency_SlotTaken(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	ready := make(chan struct{})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		<-block
		w.WriteHeader(http.StatusOK)
	})
	handler := Concurrency(next, 1)
	go handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	<-ready
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(w, r)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
