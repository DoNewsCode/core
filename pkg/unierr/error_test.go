package unierr

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"strings"
	"testing"
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

type testPrinter struct {
}

func (t testPrinter) Sprintf(msg string, val ...interface{}) string {
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
