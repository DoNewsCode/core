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

type ResponseEncoder struct {
	w http.ResponseWriter
}

func NewResponseEncoder(w http.ResponseWriter) *ResponseEncoder {
	return &ResponseEncoder{w: w}
}

func (s *ResponseEncoder) Encode(response interface{}, err error) {
	if err != nil {
		s.EncodeError(err)
		return
	}
	s.EncodeResponse(response)
}

func (s *ResponseEncoder) EncodeError(err error) {
	encode(s.w, err, http.StatusInternalServerError)
}

func (s *ResponseEncoder) EncodeResponse(response interface{}) {
	encode(s.w, response, http.StatusOK)
}

func encode(w http.ResponseWriter, any interface{}, code int) {
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

	switch any.(type) {
	case proto.Message: // gogoproto proto.Error
		marshaller := jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		}
		_ = marshaller.Marshal(w, any.(proto.Message))
	case error:
		if _, ok := any.(json.Marshaler); !ok {
			any = map[string]string{
				"error": any.(error).Error(),
			}
		}
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(any)
	default:
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(any)
	}
}
