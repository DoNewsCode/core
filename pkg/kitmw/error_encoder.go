package kitmw

import (
	"context"
	"encoding/json"
	"net/http"
)

type Headerer interface {
	Headers() http.Header
}

type StatusCoder interface {
	StatusCode() int
}

// ErrorEncoder writes the error to the ResponseWriter, by default a content
// type of application/json, a body of json with key "error" and the value
// error.Error(), and a status code of 500. If the error implements Headerer,
// the provided headers will be applied to the response. If the error
// implements json.Marshaler, and the marshaling succeeds, the JSON encoded
// form of the error will be used. If the error implements StatusCoder, the
// provided StatusCode will be used instead of 500.
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	const contentType = "application/json; charset=utf-8"
	w.Header().Set("Content-Type", contentType)

	if headerer, ok := err.(Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	code := http.StatusInternalServerError
	if sc, ok := err.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(err)
}
