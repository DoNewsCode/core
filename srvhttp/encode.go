package srvhttp

import (
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Headerer interface {
	// Headers provides the header map that will be sent by http.ResponseWriter WriteHeader.
	Headers() http.Header
}

type StatusCoder interface {
	// StatusCode provides the status code for the http response.
	StatusCode() int
}

// ResponseEncoder encodes either a successful http response or errors to the JSON format,
// and pipe the serialized json data to the http.ResponseWriter.
//
// It asserts the type of input in following order and figures out the matching encoder:
//
//  json.Marshaler: use encoding/json encoder.
//  proto.Message: use the jsonpb encoder
//  error: {"message": err.Error()}
//  by default: encoding/json encoder.
//
// It also populates http status code and headers if necessary.
type ResponseEncoder struct {
	w http.ResponseWriter
}

// NewResponseEncoder wraps the http.ResponseWriter and returns a reference to ResponseEncoder
func NewResponseEncoder(w http.ResponseWriter) *ResponseEncoder {
	return &ResponseEncoder{w: w}
}

// Encode serialize response and error to the corresponding json format and write then to the output buffer.
//
// See ResponseEncoder for details.
func (s *ResponseEncoder) Encode(response any, err error) {
	if err != nil {
		s.EncodeError(err)
		return
	}
	s.EncodeResponse(response)
}

// EncodeError encodes an Error. If the error is not a StatusCoder, the http.StatusInternalServerError will be used.
func (s *ResponseEncoder) EncodeError(err error) {
	encode(s.w, err, http.StatusInternalServerError)
}

// EncodeResponse encodes an response value.
// If the response is not a StatusCoder, the http.StatusInternalServerError will be used.
func (s *ResponseEncoder) EncodeResponse(response any) {
	encode(s.w, response, http.StatusOK)
}

func encode(w http.ResponseWriter, any any, code int) {
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

	switch x := any.(type) {
	case json.Marshaler:
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(x)
	case proto.Message:
		bytes, _ := protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}.Marshal(x)
		w.Write(bytes)
	case error:
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(map[string]string{
			"message": x.Error(),
		})
	default:
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(x)
	}
}
