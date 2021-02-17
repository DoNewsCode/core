// package mods3 contains integration with package core
package mods3

import (
	"net/http"

	"github.com/DoNewsCode/std/pkg/ots3"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Module is a s3 module that adds a file upload path. It uses the default configuration.
type Module struct {
	handler http.Handler
}

// ProvideHttp adds a "/upload" path to router
func (m Module) ProvideHttp(router *mux.Router) {
	router.PathPrefix("/upload").Handler(m.handler)
}

// New creates the s3 module
func New(manager *ots3.Manager, logger log.Logger, env contract.Env) *Module {
	uploadService := &ots3.UploadService{
		Logger: logger,
		S3:     manager,
	}
	endpoint := ots3.MakeUploadEndpoint(uploadService)
	middleware := ots3.Middleware(logger, env)
	handler := ots3.MakeHttpHandler(endpoint, middleware)
	module := &Module{
		handler: handler,
	}
	return module
}
