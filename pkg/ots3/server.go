package ots3

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/DoNewsCode/std/pkg/kitmw"
	"github.com/DoNewsCode/std/pkg/srverr"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	httptransport "github.com/go-kit/kit/transport/http"

	"io"
	"net/http"
)

type UploadService struct {
	logger log.Logger
	s3     *Manager
}

type Uploader interface {
	Upload(ctx context.Context, name string, reader io.Reader) (string, error)
}

func (s *UploadService) Upload(ctx context.Context, name string, reader io.Reader) (newUrl string, err error) {
	defer func() {
		if closer, ok := reader.(io.ReadCloser); ok {
			closer.Close()
		}
	}()
	newUrl, err = s.s3.Upload(ctx, name, reader)
	level.Info(s.logger).Log("msg", fmt.Sprintf("file %s uploaded to %s", name, newUrl))
	return newUrl, err
}

func MakeUploadEndpoint(uploader Uploader) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*Request)
		resp, err := uploader.Upload(ctx, req.name, req.data)
		if err != nil {
			return nil, srverr.InternalErr(err, "failed to upload")
		}
		return &Response{
			Code: 0,
			Data: struct {
				Url string `json:"url"`
			}{Url: resp},
		}, nil
	}
}

func Middleware(logger log.Logger, env contract.Env) endpoint.Middleware {
	keyer := key.NewKeyManager("module", "s3", "service", "upload")
	l := kitmw.MakeLoggingMiddleware(logger, keyer, env.IsLocal())
	e := kitmw.MakeErrorMarshallerMiddleware(kitmw.ErrorOption{
		AlwaysHTTP200: false,
		AlwaysGRPCOk:  false,
		ShouldRecover: env.IsProduction(),
	})
	return endpoint.Chain(e, l)
}

func MakeHttpHandler(endpoint endpoint.Endpoint, middleware endpoint.Middleware) http.Handler {
	server := httptransport.NewServer(
		middleware(endpoint),
		decodeRequest,
		httptransport.EncodeJSONResponse,
	)
	return server
}

func decodeRequest(_ context.Context, request2 *http.Request) (request interface{}, err error) {
	file, header, err := request2.FormFile("file")
	if err != nil {
		return nil, err
	}
	return &Request{
		name: header.Filename,
		data: file,
	}, nil
}
