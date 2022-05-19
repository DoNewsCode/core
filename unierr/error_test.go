package unierr

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestServerError_UnmarshalJSON(t *testing.T) {
	testError := &Error{
		err:  errors.New("err"),
		msg:  "foo",
		code: codes.Aborted,
	}
	byts, err := json.Marshal(testError)
	assert.NoError(t, err)
	var result *Error
	err = json.Unmarshal(byts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testError.code, result.code)
	assert.Equal(t, testError.msg, result.msg)
	assert.True(t, IsAbortedErr(result))
}

func TestServerError_FromStatus(t *testing.T) {
	testError := &Error{
		err:  errors.New("err"),
		msg:  "foo",
		code: codes.Aborted,
	}
	status := testError.GRPCStatus()
	assert.Equal(t, codes.Aborted, status.Code())
	assert.Equal(t, "foo", status.Message())
	result := FromStatus(status)
	assert.Equal(t, codes.Aborted, result.code)
	assert.Equal(t, "foo", result.msg)
	assert.True(t, IsAbortedErr(result))
}

type testPrinter struct{}

func (t testPrinter) Sprintf(msg string, val ...any) string {
	return strings.ToUpper(msg)
}

func TestServerError_CustomPrinter(t *testing.T) {
	testError := &Error{
		err:     errors.New("err"),
		msg:     "foo",
		code:    codes.Aborted,
		Printer: testPrinter{},
	}
	status := testError.GRPCStatus()
	assert.Equal(t, codes.Aborted, status.Code())
	assert.Equal(t, "FOO", status.Message())
	bytes, err := testError.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"code":10,"message":"FOO"}`), bytes)
}

func TestWrap(t *testing.T) {
	type args struct {
		err  error
		code codes.Code
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"err_nil", args{nil, codes.Aborted}, codes.Aborted.String()},
		{"err_foo", args{errors.New("foo"), codes.Aborted}, "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testError := Wrap(tt.args.err, tt.args.code)
			assert.Equal(t, tt.want, testError.Error())
			byts, err := json.Marshal(testError)
			assert.NoError(t, err)
			var result *Error
			err = json.Unmarshal(byts, &result)
			assert.NoError(t, err)
			assert.Equal(t, testError.code, result.code)
			assert.Equal(t, testError.msg, result.msg)
			assert.True(t, IsAbortedErr(result))

			status := testError.GRPCStatus()
			assert.Equal(t, codes.Aborted, status.Code())
			assert.Equal(t, testError.Error(), status.Message())
		})
	}
}

func TestError_StackTrace(t *testing.T) {
	type args struct {
		err  error
		code codes.Code
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"err_nil", args{nil, codes.Aborted}, 0},
		{"err_foo", args{errors.New("foo"), codes.Aborted}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{
				err:  tt.args.err,
				code: tt.args.code,
			}
			s := e.StackTrace()
			assert.Equal(t, tt.want, len(s))
		})
	}
}
