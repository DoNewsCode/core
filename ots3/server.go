package ots3

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/kitmw"
	"github.com/DoNewsCode/core/unierr"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	httptransport "github.com/go-kit/kit/transport/http"

	"io"
	"net/http"
)

// An UploadService is a go kit service to handle file upload
type UploadService struct {
	Logger log.Logger
	S3     *Manager
}

// Uploader models UploadService
type Uploader interface {
	// Upload the bytes from io.Reader with a given filename to a server, and returns the url and error.
	Upload(ctx context.Context, name string, reader io.Reader) (string, error)
}

// Upload reads all bytes from reader, and upload then to S3 as the provided name. The url will be returned.
func (s *UploadService) Upload(ctx context.Context, name string, reader io.Reader) (newUrl string, err error) {
	defer func() {
		if closer, ok := reader.(io.ReadCloser); ok {
			closer.Close()
		}
	}()
	newUrl, err = s.S3.Upload(ctx, name, reader)
	level.Info(s.Logger).Log("msg", fmt.Sprintf("file %s uploaded to %s", name, newUrl))
	return newUrl, err
}

// MakeUploadEndpoint creates a Upload endpoint
func MakeUploadEndpoint(uploader Uploader) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*Request)
		resp, err := uploader.Upload(ctx, req.name, req.data)
		if err != nil {
			return nil, unierr.InternalErr(err, "failed to upload")
		}
		return &Response{
			Code: 0,
			Data: struct {
				Url string `json:"url"`
			}{Url: resp},
		}, nil
	}
}

// Middleware adds logging and error handling to the endpoint.
func Middleware(logger log.Logger, env contract.Env) endpoint.Middleware {
	keyer := key.New("module", "S3", "service", "upload")
	l := kitmw.MakeLoggingMiddleware(logger, keyer, env.IsLocal())
	e := kitmw.MakeErrorConversionMiddleware(kitmw.ErrorOption{
		AlwaysHTTP200: false,
		ShouldRecover: env.IsProduction(),
	})
	return endpoint.Chain(e, l)
}

// MakeHttpHandler creates a go kit transport in http for *UploadService.
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
