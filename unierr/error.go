// Package unierr presents an unification error model between gRPC transport and HTTP transport,
// and between server and client.
//
// It is modeled after the gRPC status.
//
// To create an not found error with a custom message:
//
//  unierr.New(codes.NotFound, "some stuff is missing")
//
// To wrap an existing error:
//
//  unierr.Wrap(err, codes.NotFound)
//
// See example for detailed usage.
package unierr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/text"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New returns an error representing code and msg. If code is OK, returns nil.
func New(code codes.Code, msg string) *Error {
	if code == codes.OK {
		return nil
	}
	return &Error{
		msg:  msg,
		code: code,
	}
}

// Newf returns New(code, fmt.Sprintf(format, args...)).
func Newf(code codes.Code, format string, args ...interface{}) *Error {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap annotates an error with a codes.Code
func Wrap(err error, code codes.Code) *Error {
	err = errors.WithStack(err)
	return &Error{
		err:  err,
		code: code,
		msg:  err.Error(),
	}
}

// Wrapf annotates an error with a codes.Code, and provides a new error message.
// The wrapped error hence is mainly kept for tracing and debugging. The message
// in the wrapped error becomes irrelevant as it is overwritten by the new
// message.
func Wrapf(err error, code codes.Code, format string, args ...interface{}) *Error {
	se := Wrap(err, code)
	se.msg = format
	se.args = args
	return se
}

// Error is the unified error type for HTTP/gRPC transports.
// In grpc transports, Error can not only be constructed from a grpc status but also producing a native grpc status.
// In HTTP transports, Error can be encoded and decoded in json format. It also infers HTTP status code.
//
// The roundtrip conversion makes Error suitable as a unification error model, on both client side and server side.
// Note the json format follows the JSONRPC standard.
type Error struct {
	err  error
	msg  string
	args []interface{}
	code codes.Code
	// Printer can ben used to achieve i18n. By default it is a text.BasePrinter.
	Printer contract.Printer
	// HttpStatusCodeFunc can overwrites the inferred HTTP status code from gRPC status.
	HttpStatusCodeFunc func(code codes.Code) int
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Error) UnmarshalJSON(bytes []byte) error {
	var jsonRepresentation struct {
		Code  uint32 `json:"code"`
		Error string `json:"message"`
	}
	if err := json.Unmarshal(bytes, &jsonRepresentation); err != nil {
		return err
	}
	e.code = codes.Code(jsonRepresentation.Code)
	e.msg = jsonRepresentation.Error
	e.err = errors.New(e.msg)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (e *Error) MarshalJSON() (result []byte, err error) {
	var jsonRepresentation struct {
		Code  uint32 `json:"code,omitempty"`
		Error string `json:"message"`
	}
	jsonRepresentation.Code = uint32(e.code)
	jsonRepresentation.Error = e.Error()
	return json.Marshal(jsonRepresentation)
}

// Error implements error. it consults the Printer for the output.
func (e *Error) Error() string {
	if e.Printer == nil {
		e.Printer = text.BasePrinter{}
	}
	return e.Printer.Sprintf(e.msg, e.args...)
}

// GRPCStatus produces a native gRPC status.
func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.code, e.Error())
}

// FromStatus constructs the Error from a gRPC status.
func FromStatus(s *status.Status) *Error {
	return &Error{
		err:  s.Err(),
		msg:  s.Message(),
		code: s.Code(),
	}
}

// StatusCode infers the correct http status corresponding to Error's internal code.
// If a HttpStatusCode is set in Error, that status code will be used instead.
func (e *Error) StatusCode() int {
	if e.HttpStatusCodeFunc != nil {
		return e.HttpStatusCodeFunc(e.code)
	}
	switch e.code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return 500
	}
}

// Unwrap implements go's standard errors.Unwrap() interface
func (e *Error) Unwrap() error {
	return e.err
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StackTrace implements the interface of errors.Wrap()
func (e *Error) StackTrace() errors.StackTrace {
	if err, ok := e.err.(stackTracer); ok {
		return err.StackTrace()
	}
	return errors.WithStack(e.err).(stackTracer).StackTrace()
}

// UnknownErr creates an Error with codes.Unknown.
func UnknownErr(e error) *Error {
	return err(codes.Unknown, e)
}

// IsUnknownErr checks if an Error has codes.Unknown
func IsUnknownErr(e error) bool {
	return is(e, codes.Unknown)
}

// CanceledErr creates an Error with codes.Canceled
func CanceledErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Canceled, e, msgAndArgs...)
}

// IsCanceledErr checks if an Error has codes.Canceled
func IsCanceledErr(e error) bool {
	return is(e, codes.Canceled)
}

// DeadlineExceededErr creates an Error with codes.DeadlineExceeded
func DeadlineExceededErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.DeadlineExceeded, e, msgAndArgs...)
}

// IsDeadlineExceededErr checks if an Error has codes.DeadlineExceeded
func IsDeadlineExceededErr(e error) bool {
	return is(e, codes.DeadlineExceeded)
}

// AlreadyExistsErr creates an Error with codes.AlreadyExists
func AlreadyExistsErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.AlreadyExists, e, msgAndArgs...)
}

// IsAlreadyExistsErr checks if an Error has codes.AlreadyExists
func IsAlreadyExistsErr(e error) bool {
	return is(e, codes.AlreadyExists)
}

// AbortedErr creates an Error with codes.Aborted
func AbortedErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Aborted, e, msgAndArgs...)
}

// IsAbortedErr checks if an Error has codes.Aborted
func IsAbortedErr(e error) bool {
	return is(e, codes.Aborted)
}

// OutOfRangeErr creates an Error with codes.OutOfRange
func OutOfRangeErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.OutOfRange, e, msgAndArgs...)
}

// IsOutOfRangeErr checks if an Error has codes.OutOfRange
func IsOutOfRangeErr(e error) bool {
	return is(e, codes.OutOfRange)
}

// UnimplementedErr creates an Error with codes.Unimplemented
func UnimplementedErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Unimplemented, e, msgAndArgs...)
}

// IsUnimplementedErr checks if an Error has codes.Unimplemented
func IsUnimplementedErr(e error) bool {
	return is(e, codes.Unimplemented)
}

// InternalErr creates an Error with codes.Internal
func InternalErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Internal, e, msgAndArgs...)
}

// IsInternalErr checks if an Error has codes.Internal
func IsInternalErr(e error) bool {
	return is(e, codes.Internal)
}

// PermissionDeniedErr creates an Error with codes.PermissionDenied
func PermissionDeniedErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.PermissionDenied, e, msgAndArgs...)
}

// IsPermissionDeniedErr checks if an Error has codes.PermissionDenied
func IsPermissionDeniedErr(e error) bool {
	return is(e, codes.PermissionDenied)
}

// InvalidArgumentErr creates an Error with codes.InvalidArgument
func InvalidArgumentErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.InvalidArgument, e, msgAndArgs...)
}

// IsInvalidArgumentErr checks if an Error has codes.InvalidArgument
func IsInvalidArgumentErr(e error) bool {
	return is(e, codes.InvalidArgument)
}

// NotFoundErr creates an Error with codes.NotFound
func NotFoundErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.NotFound, e, msgAndArgs...)
}

// IsNotFoundErr checks if an Error has codes.NotFound
func IsNotFoundErr(e error) bool {
	return is(e, codes.NotFound)
}

// UnavailableErr creates an Error with codes.Unavailable
func UnavailableErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Unavailable, e, msgAndArgs...)
}

// IsUnavailableErr checks if an Error has codes.Unavailable
func IsUnavailableErr(e error) bool {
	return is(e, codes.Unavailable)
}

// DataLossErr creates an Error with codes.DataLoss
func DataLossErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.DataLoss, e, msgAndArgs...)
}

// IsDataLossErr checks if an Error has codes.DataLoss
func IsDataLossErr(e error) bool {
	return is(e, codes.DataLoss)
}

// UnauthenticatedErr creates an Error with codes.Unauthenticated
func UnauthenticatedErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.Unauthenticated, e, msgAndArgs...)
}

// IsUnauthenticatedErr checks if an Error has codes.Unauthenticated
func IsUnauthenticatedErr(e error) bool {
	return is(e, codes.Unauthenticated)
}

// ResourceExhaustedErr creates an Error with codes.ResourceExhausted
func ResourceExhaustedErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.ResourceExhausted, e, msgAndArgs...)
}

// IsResourceExhaustedErr checks if an Error has codes.ResourceExhausted
func IsResourceExhaustedErr(e error) bool {
	return is(e, codes.ResourceExhausted)
}

// FailedPreconditionErr creates an Error with codes.FailedPrecondition
func FailedPreconditionErr(e error, msgAndArgs ...interface{}) *Error {
	return err(codes.FailedPrecondition, e, msgAndArgs...)
}

// IsFailedPreconditionErr checks if an Error has codes.FailedPrecondition
func IsFailedPreconditionErr(e error) bool {
	return is(e, codes.FailedPrecondition)
}

func err(code codes.Code, e error, msgAndArgs ...interface{}) *Error {
	if len(msgAndArgs) == 0 {
		return Wrap(e, code)
	}
	if s, ok := msgAndArgs[0].(string); ok {
		return Wrapf(e, code, s, msgAndArgs[1:]...)
	}
	return Wrapf(e, code, e.Error(), msgAndArgs...)
}

func is(err error, code codes.Code) bool {
	var serverError *Error
	if errors.As(err, &serverError) {
		return serverError.code == code
	}
	return false
}
