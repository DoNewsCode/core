package ots3

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type Request struct {
	name string
	data io.Reader
}

type Response struct {
	Data struct {
		Url string `json:"url"`
	} `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewClient(uri *url.URL) *httptransport.Client {
	return httptransport.NewClient("POST", uri, encodeClientRequest, decodeClientResponse)
}

func decodeClientResponse(_ context.Context, response2 *http.Response) (response interface{}, err error) {
	defer response2.Body.Close()
	b, err := ioutil.ReadAll(response2.Body)
	if err != nil {
		return nil, err
	}
	var resp Response
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func encodeClientRequest(ctx context.Context, request *http.Request, i interface{}) error {
	defer func() {
		if rc, ok := i.(Request).data.(io.ReadCloser); ok {
			rc.Close()
		}
	}()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", i.(Request).name)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, i.(Request).data)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Body = ioutil.NopCloser(body)
	return nil
}

type ClientUploader struct {
	endpoint endpoint.Endpoint
}

func (c ClientUploader) Upload(ctx context.Context, name string, reader io.Reader) (newUrl string, err error) {
	resp, err := c.endpoint(ctx, Request{data: reader, name: name})
	if err != nil {
		return "", err
	}
	return resp.(Response).Data.Url, err
}

func NewClientUploader(client *httptransport.Client) *ClientUploader {
	return &ClientUploader{endpoint: client.Endpoint()}
}

func InjectClientUploader(uri *url.URL) *ClientUploader {
	client := NewClient(uri)
	clientUploader := NewClientUploader(client)
	return clientUploader
}
