package ots3

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	t.Parallel()
	manager := setupManager()
	service := UploadService{
		Logger: log.NewNopLogger(),
		S3:     manager,
	}
	endpoint := MakeUploadEndpoint(&service)
	endpoint = Middleware(log.NewNopLogger(), config.Env("testing"))(endpoint)
	handler := httptransport.NewServer(endpoint, decodeRequest, httptransport.EncodeJSONResponse)
	ln, _ := net.Listen("tcp", ":8888")
	server := &http.Server{
		Handler: handler,
	}
	go server.Serve(ln)
	defer server.Shutdown(context.Background())

	uri, _ := url.Parse("http://localhost:8888/")
	uploader := NewClientUploaderFromUrl(uri)
	urlStr, err := uploader.Upload(context.Background(), "foo", strings.NewReader("bar"))
	assert.NoError(t, err)
	assert.NotEmpty(t, urlStr)
}
