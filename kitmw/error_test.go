package kitmw

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/DoNewsCode/core/unierr"
	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestMakeErrorMarshallerMiddleware(t *testing.T) {
	mw := Error(ErrorOption{
		AlwaysHTTP200: false,
		ShouldRecover: false,
	})
	e1 := func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, errors.New("foo")
	}
	e2 := func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, unierr.NotFoundErr(errors.New("bar"), "")
	}
	e3 := func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, unierr.NotFoundErr(unierr.InvalidArgumentErr(errors.New("bar"), ""), "")
	}
	e4 := func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, errors.Wrap(unierr.NotFoundErr(errors.New("foo"), ""), "bar")
	}
	cases := []endpoint.Endpoint{e1, e2, e3, e4}
	for _, c := range cases {
		cc := c
		t.Run("", func(t *testing.T) {
			_, err := mw(cc)(nil, nil)
			if _, ok := err.(*unierr.Error); !ok {
				t.Fail()
			}
		})
	}
}

func TestPanicRecover(t *testing.T) {
	mw := Error(ErrorOption{
		AlwaysHTTP200: false,
		ShouldRecover: true,
	})
	e1 := func(ctx context.Context, request interface{}) (interface{}, error) {
		panic("test")
	}
	cases := []endpoint.Endpoint{e1}
	for _, c := range cases {
		cc := c
		t.Run("", func(t *testing.T) {
			_, err := mw(cc)(nil, nil)
			if _, ok := err.(*unierr.Error); !ok {
				t.Fail()
			}
		})
	}
}

func TestErrorEncoder(t *testing.T) {
	var err = unierr.InternalErr(errors.New("server bug"), "whoops")
	recorder := httptest.NewRecorder()
	ErrorEncoder(context.Background(), err, recorder)
	resp := recorder.Result()
	assert.Equal(t, "500 Internal Server Error", resp.Status)
	content, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "{\"code\":13,\"message\":\"whoops\"}\n", string(content))
}
