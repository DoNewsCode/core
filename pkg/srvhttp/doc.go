package srvhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AddDocMiddleware returns a documentation path at /doc/
func MakeDocMiddleware() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		router := mux.NewRouter()
		router.PathPrefix("/docs/").Handler(getOpenAPIHandler())
		router.PathPrefix("/docs").Handler(http.RedirectHandler("/docs/", 302))
		router.PathPrefix("/").Handler(handler)
		return router
	}
}

func Doc(router *mux.Router) {
	router.PathPrefix("/docs/").Handler(getOpenAPIHandler())
	router.PathPrefix("/docs").Handler(http.RedirectHandler("/docs/", 302))
}

// getOpenAPIHandler serves an OpenAPI UI.
func getOpenAPIHandler() http.Handler {
	return http.StripPrefix("/docs", http.FileServer(http.Dir("./docs")))
}
