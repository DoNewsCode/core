package srverr

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/text"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func err(code codes.Code, e error, msgAndArgs ...interface{}) ServerError {
	if len(msgAndArgs) == 0 {
		return ServerError{err: e, msg: e.Error(), customCode: uint32(code)}
	}
	if s, ok := msgAndArgs[0].(string); ok {
		return ServerError{err: e, msg: s, args: msgAndArgs[1:], customCode: uint32(code)}
	}
	return ServerError{err: e, msg: "%+v", args: msgAndArgs, customCode: uint32(code)}
}

func UnknownErr(e error) ServerError {
	return err(codes.Unknown, e, redact(e))
}

func CanceledErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Canceled, e, msgAndArgs...)
}

func DeadlineExceededErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.DeadlineExceeded, e, msgAndArgs...)
}

func AlreadyExistsErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.AlreadyExists, e, msgAndArgs...)
}

func AbortedErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Aborted, e, msgAndArgs...)
}

func OutOfRangeErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.OutOfRange, e, msgAndArgs...)
}

func UnimplementedErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Unimplemented, e, msgAndArgs...)
}

func InternalErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Unimplemented, e, msgAndArgs...)
}

func PermissionDeniedErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.PermissionDenied, e, msgAndArgs...)
}

func InvalidArgumentErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.InvalidArgument, e, msgAndArgs...)
}

func NotFoundErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.NotFound, e, msgAndArgs...)
}

func UnavailableErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Unavailable, e, msgAndArgs...)
}

func DataLossErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.DataLoss, e, msgAndArgs...)
}

func UnauthenticatedErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.Unauthenticated, e, msgAndArgs...)
}

func ResourceExhaustedErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.ResourceExhausted, e, msgAndArgs...)
}

func FailedPreconditionErr(e error, msgAndArgs ...interface{}) ServerError {
	return err(codes.FailedPrecondition, e, msgAndArgs...)
}

func redact(err error) string {
	return strings.Split(err.Error(), ":")[0]
}

type ServerError struct {
	err        error
	msg        string
	args       []interface{}
	customCode uint32
	Printer    contract.Printer
	HttpStatusCode int
	GrpcStatusCode int
}

func (e ServerError) MarshalJSON() ([]byte, error) {
	type jsonRep struct {
		Code    uint32 `json:"code"`
		Msg     string `json:"msg"`
	}
	if e.Printer == nil {
		e.Printer = text.BasePrinter{}
	}
	r := jsonRep{
		e.customCode,
		e.Printer.Sprintf(e.msg, e.args...),
	}
	return json.Marshal(r)
}

func (e ServerError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.msg
}

func (e ServerError) GRPCStatus() *status.Status {
	if e.GrpcStatusCode != 0 {
		return status.New(codes.Code(e.GrpcStatusCode), e.msg)
	}
	if e.customCode >= 17 {
		return status.New(codes.Unknown, e.msg)
	}
	return status.New(codes.Code(e.customCode), e.msg)
}

// StatusCode Implements https status
func (e ServerError) StatusCode() int {
	if e.HttpStatusCode != 0 {
		return e.HttpStatusCode
	}
	switch codes.Code(e.customCode) {
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
func (e ServerError) Unwrap() error {
	return e.err
}

// StackTrace implements the interface of errors.Wrap()
func (e ServerError) StackTrace() errors.StackTrace {
	if err, ok := e.err.(stackTracer); ok {
		return err.StackTrace()
	}
	return errors.WithStack(e.err).(stackTracer).StackTrace()
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

