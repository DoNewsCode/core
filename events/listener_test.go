package events

import (
	"context"
	"testing"

	"github.com/DoNewsCode/core/contract"
	"github.com/stretchr/testify/assert"
)

type mockListener1 struct{}

func (m mockListener1) Listen() (topic interface{}) {
	panic("implement me")
}

func (m mockListener1) Process(ctx context.Context, payload interface{}) error {
	panic("implement me")
}

func (m mockListener1) Equals(listener contract.Listener) bool {
	return true
}

type mockListener2 struct{}

func (m mockListener2) Listen() (topic interface{}) {
	panic("implement me")
}

func (m mockListener2) Process(ctx context.Context, payload interface{}) error {
	panic("implement me")
}

func (m mockListener2) Equals(listener contract.Listener) bool {
	return false
}

func TestOnceListener(t *testing.T) {
	once := onceListener{nil, mockListener1{}}
	assert.True(t, once.Equals(mockListener1{}))

	once = onceListener{nil, mockListener2{}}
	assert.False(t, once.Equals(mockListener2{}))
}
