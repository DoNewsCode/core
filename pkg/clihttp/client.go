package clihttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type Client struct {
	tracer               opentracing.Tracer
	underlying           contract.HttpDoer
	requestLogThreshold  int
	responseLogThreshold int
}

type option func(*Client)

func WithDoer(doer contract.HttpDoer) option {
	return func(client *Client) {
		client.underlying = doer
	}
}

func WithRequestLogThreshold(num int) option {
	return func(client *Client) {
		client.requestLogThreshold = num
	}
}

func WithResponseLogThreshold(num int) option {
	return func(client *Client) {
		client.requestLogThreshold = num
	}
}

func NewClient(tracer opentracing.Tracer, options ...option) *Client {
	baseClient := &http.Client{Transport: &nethttp.Transport{}}
	c := &Client{
		tracer: tracer,
		underlying: baseClient,
		requestLogThreshold: 5000,
		responseLogThreshold: 5000,
	}
	for _, f := range options {
		f(c)
	}
	return c
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req, tracer := nethttp.TraceRequest(c.tracer, req)
	defer tracer.Finish()

	c.logRequest(req, tracer)

	response, err := c.underlying.Do(req)
	if err != nil {
		return response, err
	}

	c.logResponse(response, tracer)

	return response, err
}

func (c *Client) logRequest(req *http.Request, tracer *nethttp.Tracer) {
	if req.Body == nil {
		return
	}
	body, err := req.GetBody()
	if err != nil {
		tracer.Span().LogKV("error", errors.Wrap(err, "cannot get request body"))
		return
	}
	length, _ := strconv.Atoi(req.Header.Get(http.CanonicalHeaderKey("Content-Length")))
	if length > c.requestLogThreshold {
		tracer.Span().LogKV("request", "elided: Content-Length too large")
		return
	}
	byt, err := ioutil.ReadAll(body)
	if err != nil {
		tracer.Span().LogKV("error", errors.Wrap(err, "cannot read request body"))
		return
	}
	tracer.Span().LogKV("request", string(byt))
}

func (c *Client) logResponse(response *http.Response, tracer *nethttp.Tracer) {
	length, _ := strconv.Atoi(response.Header.Get(http.CanonicalHeaderKey("Content-Length")))
	if length > c.responseLogThreshold {
		tracer.Span().LogKV("response", "elided: Content-Length too large")
		return
	}
	var buf bytes.Buffer
	byt, err := ioutil.ReadAll(response.Body)
	if err != nil {
		tracer.Span().LogKV("error", errors.Wrap(err, "cannot read response body"))
	}
	tracer.Span().LogKV("response", string(byt))
	buf.Write(byt)
	response.Body = ioutil.NopCloser(&buf)
}
