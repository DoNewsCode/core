package srvhttp

import (
	"bytes"
	"github.com/DoNewsCode/std/pkg/unierr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type MockWriter struct {
	code   int
	header http.Header
	buffer bytes.Buffer
}

func (m *MockWriter) Header() http.Header {
	return m.header
}

func (m *MockWriter) Write(bytes []byte) (int, error) {
	return m.buffer.Write(bytes)
}

func (m *MockWriter) WriteHeader(statusCode int) {
	m.code = statusCode
}

func TestEncoder(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		error    error
		expected *MockWriter
	}{
		{
			"normal struct",
			struct {
				Foo string `json:"foo"`
			}{"foo"},
			nil,
			&MockWriter{
				code:   200,
				header: make(http.Header),
				buffer: *bytes.NewBufferString(`{"foo":"foo"}` + "\n"),
			},
		},
		{
			"empty string",
			nil,
			nil,
			&MockWriter{
				code:   200,
				header: make(http.Header),
				buffer: *bytes.NewBufferString(`null` + "\n"),
			},
		},
		{
			"error response",
			nil,
			errors.New("foo"),
			&MockWriter{
				code:   500,
				header: make(http.Header),
				buffer: *bytes.NewBufferString(`{"message":"foo"}` + "\n"),
			},
		},
		{
			"Error response",
			nil,
			unierr.NotFoundErr(errors.New("foo"), "bar"),
			&MockWriter{
				code:   404,
				header: make(http.Header),
				buffer: *bytes.NewBufferString(`{"code":5,"message":"bar"}` + "\n"),
			},
		},
	}
	for _, cc := range cases {
		c := cc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			writer := &MockWriter{header: make(http.Header)}
			NewResponseEncoder(writer).Encode(c.input, c.error)
			assert.Equal(t, c.expected.code, writer.code)
			assert.Equal(t, c.expected.buffer.String(), writer.buffer.String())
		})
	}
}
