package srvhttp

import (
	"encoding/json"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
)

type Headerer interface {
	Headers() http.Header
}

type StatusCoder interface {
	StatusCode() int
}

type ServiceEncoder struct {
	w http.ResponseWriter
}

func NewEncoder(w http.ResponseWriter) ServiceEncoder {
	return ServiceEncoder{w: w}
}

func (s ServiceEncoder) Encode(response interface{}, err error) {
	if err != nil {
		EncodeError(s.w, err)
		return
	}
	EncodeResponse(s.w, response)
}

func EncodeResponse(w http.ResponseWriter, response interface{}) {
	encodeGeneric(w, response, http.StatusOK)
}

func EncodeError(w http.ResponseWriter, err error) {
	encodeGeneric(w, err, http.StatusInternalServerError)
}

func encodeGeneric(w http.ResponseWriter, any interface{}, code int) {
	const contentType = "application/json; charset=utf-8"
	w.Header().Set("Content-Type", contentType)

	if headerer, ok := any.(Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	if sc, ok := any.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	// Pick the right encoder
	switch any.(type) {
	case proto.Message:
		marshaller := jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		}
		_ = marshaller.Marshal(w, any.(proto.Message))
	default:
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(any)
	}
}
