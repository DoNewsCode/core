package kitmw

import (
	"context"
	"github.com/DoNewsCode/core/srvhttp"
	"net/http"
)

// ErrorEncoder is a go kit style http error encoder. Internally it uses
// srvhttp.ResponseEncoder to encode the error.
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	encoder := srvhttp.NewResponseEncoder(w)
	encoder.EncodeError(err)
}
