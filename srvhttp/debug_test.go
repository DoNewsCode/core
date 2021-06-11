package srvhttp

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebugModule(t *testing.T) {
	router := mux.NewRouter()
	DebugModule{}.ProvideHTTP(router)

	paths := []string{
		"/debug/pprof/cmdline",
		"/debug/pprof/profile?seconds=1",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
		"/debug/pprof/heap",
		"/debug/pprof/allocs",
		"/debug/pprof/block",
		"/debug/pprof/goroutine",
		"/debug/pprof/mutex",
		"/debug/pprof/threadcreate",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}
