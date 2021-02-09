package srverr

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerError_UnmarshalJSON(t *testing.T) {
	testError := ServerError{
		err:        errors.New("err"),
		msg:        "foo",
		customCode: 42,
	}
	byts, err := json.Marshal(testError)
	assert.NoError(t, err)
	var result ServerError
	err = json.Unmarshal(byts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testError.customCode, result.customCode)
	assert.Equal(t, testError.msg, result.msg)
	assert.NotNil(t, result.err)
}
