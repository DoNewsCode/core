package core

import "github.com/gorilla/mux"

// HttpFunc converts a function to a module provides http.
type HttpFunc func(router *mux.Router)

// ProvideHTTP implements container.HTTPProvider
func (h HttpFunc) ProvideHTTP(router *mux.Router) {
	h(router)
}
