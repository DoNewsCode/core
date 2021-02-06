package ots3

import (
	"net/http"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type Module struct {
	handler http.Handler
}

func (m Module) ProvideHttp(router *mux.Router) {
	router.PathPrefix("/upload").Handler(m.handler)
}

func New(conf contract.ConfigAccessor, logger log.Logger) *Module {
	return injectModule(conf, logger)
}

func injectModule(conf contract.ConfigAccessor, logger log.Logger) *Module {
	manager, _, _ := ProvideUploadManager(S3Param{Conf: conf})
	uploadService := &UploadService{
		logger: logger,
		s3:     manager,
	}
	endpoint := MakeUploadEndpoint(uploadService)
	var envStr string
	conf.Unmarshal("env", &envStr)
	env := config.NewEnv(envStr)
	middleware := Middleware(logger, env)
	handler := MakeHttpHandler(endpoint, middleware)
	module := &Module{
		handler: handler,
	}
	return module
}
