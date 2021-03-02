package ots3

import (
	"net/http"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Module is a s3 module that adds a file upload path. It uses the default configuration.
type Module struct {
	handler http.Handler
}

// ProvideHttp adds a "/upload" path to router
func (m Module) ProvideHTTP(router *mux.Router) {
	router.PathPrefix("/upload").Handler(m.handler)
}

// New creates the s3 module
func New(manager *Manager, logger log.Logger, env contract.Env) Module {
	uploadService := &UploadService{
		Logger: logger,
		S3:     manager,
	}
	endpoint := MakeUploadEndpoint(uploadService)
	middleware := Middleware(logger, env)
	handler := MakeHttpHandler(endpoint, middleware)
	module := Module{
		handler: handler,
	}
	return module
}
