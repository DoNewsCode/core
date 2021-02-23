package ots3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/stretchr/testify/assert"
)

type mockClient func(req *http.Request) (*http.Response, error)

func (m mockClient) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

func TestNewClient(t *testing.T) {
	uri, _ := url.Parse("http://localhost/")
	client := NewClient(uri)
	assert.NotNil(t, client)
}

func TestClientUploader_Upload(t *testing.T) {
	var errUploaded = errors.New("testing")
	cases := []struct {
		name    string
		client  httptransport.HTTPClient
		asserts func(t *testing.T, url string, err error)
	}{
		{
			"error result",
			mockClient(func(req *http.Request) (*http.Response, error) {
				f, h, _ := req.FormFile("file")
				fileContent, _ := ioutil.ReadAll(f)
				assert.Equal(t, "foo", h.Filename)
				assert.Equal(t, fileContent, []byte("bar"))
				return nil, errUploaded
			}),
			func(t *testing.T, url string, err error) {
				assert.Equal(t, url, "")
				assert.ErrorIs(t, err, errUploaded)
			},
		},
		{
			"success result",
			mockClient(func(req *http.Request) (*http.Response, error) {
				f, h, _ := req.FormFile("file")
				fileContent, _ := ioutil.ReadAll(f)
				assert.Equal(t, "foo", h.Filename)
				assert.Equal(t, fileContent, []byte("bar"))

				var (
					resp http.Response
					buf  bytes.Buffer
				)
				result := map[string]interface{}{
					"code": 0,
					"data": map[string]string{
						"url": "http://donews.com",
					},
				}
				json.NewEncoder(&buf).Encode(result)
				resp.Body = ioutil.NopCloser(&buf)
				resp.StatusCode = 200
				return &resp, nil
			}),
			func(t *testing.T, url string, err error) {
				assert.Equal(t, url, "http://donews.com")
				assert.NoError(t, err)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			uri, _ := url.Parse("http://localhost/")
			client := httptransport.NewClient(
				"POST",
				uri,
				encodeClientRequest,
				decodeClientResponse,
				httptransport.SetClient(c.client),
			)
			uploader := NewClientUploader(client)
			urlStr, err := uploader.Upload(context.Background(), "foo", strings.NewReader("bar"))
			c.asserts(t, urlStr, err)
		})
	}
}
