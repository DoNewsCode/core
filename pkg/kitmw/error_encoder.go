package kitmw

import (
	"context"
	"github.com/DoNewsCode/std/pkg/srvhttp"
	"net/http"
)

// ErrorEncoder writes the error to the ResponseWriter, by default a content
// type of application/json, a body of json with key "error" and the value
// error.Error(), and a status code of 500. If the error implements Headerer,
// the provided headers will be applied to the response. If the error
// implements json.Marshaler, and the marshaling succeeds, the JSON encoded
// form of the error will be used. If the error implements StatusCoder, the
// provided StatusCode will be used instead of 500.
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	encoder := srvhttp.NewResponseEncoder(w)
	encoder.EncodeError(err)
}
