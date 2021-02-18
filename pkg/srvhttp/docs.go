package srvhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

// DocsModule defines a http provider for container.Container. It serves static files under `./docs`.
// `./docs` is supposed to contain doc serving engines such as Swagger, Docify or plain markdowns.
type DocsModule struct{}

// ProvideHttp implements container.HttpProvider
func (d DocsModule) ProvideHttp(router *mux.Router) {
	router.PathPrefix("/docs/").Handler(serveStatic())
	router.PathPrefix("/docs").Handler(http.RedirectHandler("/docs/", 302))
}

func serveStatic() http.Handler {
	// in Go 1.16 go:embed can be used instead of FileServer.
	return http.StripPrefix("/docs", http.FileServer(http.Dir("./docs")))
}
